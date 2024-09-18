package message

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/sensdata/idb/core/encrypt"
)

// 消息起始字节
const (
	MagicBytes    = "\xAB\xCD\xEF\x01"
	MagicBytes1   = "\xAB\xCD\xEF\x02"
	MagicBytesLen = 4
	MsgLenBytes   = 4
)

// 消息类型
type MessageType string
type FileMessageType string

const (
	Heartbeat     MessageType = "hb"
	CmdMessage    MessageType = "cmd"
	ActionMessage MessageType = "action"
)

const (
	Upload   FileMessageType = "upload"
	Download FileMessageType = "download"
)

const (
	FileOk  int = 0
	FileErr int = 1
)

// 消息数据分隔符
const Separator string = "#idb#"

type MessageInterface interface {
	GetType() string
}

// 消息体
type Message struct {
	MsgID     string      `json:"msg_id"`
	Type      MessageType `json:"type"`
	Sign      string      `json:"sign"`
	Data      string      `json:"data"`
	Timestamp int64       `json:"timestamp"`
	Nonce     string      `json:"nonce"`
	Version   string      `json:"version"`
	Checksum  string      `json:"checksum"`
}

func (m *Message) GetType() string {
	return "Message"
}

// 文件消息
type FileMessage struct {
	MsgID     string          `json:"msg_id"`
	Type      FileMessageType `json:"type"`
	Status    int             `json:"status"`                        // 状态
	Path      string          `json:"path" validate:"required"`      // 文件路径
	FileName  string          `json:"file_name" validate:"required"` //文件名
	TotalSize int64           `json:"total_size"`                    // 文件总大小（可选，Agent端可校验完整性）
	Offset    int64           `json:"offset"`                        // 当前文件块的起始偏移量
	ChunkSize int             `json:"chunk_size"`                    // 当前文件块的大小
	Chunk     []byte          `json:"chunk"`                         // 当前文件块
}

func (f *FileMessage) GetType() string {
	return "FileMessage"
}

// ErrIncompleteMessage 表示接收到的消息数据不完整
var ErrIncompleteMessage = errors.New("incomplete message data")

func CreateFileMessage(msgID string, msgType FileMessageType, status int, path string, name string, total int64, offset int64, chunkSize int, chunk []byte) (*FileMessage, error) {
	// 构造要发送的消息
	fileMessage := &FileMessage{
		MsgID:     msgID,
		Type:      msgType,
		Status:    status,
		Path:      path,
		FileName:  name,
		TotalSize: total,
		Offset:    offset,
		ChunkSize: chunkSize,
		Chunk:     chunk[:chunkSize],
	}

	return fileMessage, nil
}

func SendFileMessage(conn net.Conn, msg *FileMessage) error {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 将消息长度编码到前 4 个字节
	msgLen := make([]byte, MsgLenBytes)
	binary.BigEndian.PutUint32(msgLen, uint32(len(data)))

	// 拼接消息
	encodedMsg := append([]byte(MagicBytes1), msgLen...)
	encodedMsg = append(encodedMsg, data...)

	fmt.Printf("Send:\n")
	fmt.Println(hex.EncodeToString(encodedMsg))
	fmt.Println()

	// 发送魔术字节、消息头和消息体
	_, err = conn.Write(encodedMsg)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// CreateMessage 创建并签名一个消息
func CreateMessage(msgID string, data string, key string, nonce string, msgType MessageType) (*Message, error) {
	// 时间戳
	timestamp := time.Now().Unix()

	// 加密
	encryptedData, err := encrypt.Encrypt(data, key)
	if err != nil {
		return nil, err
	}

	// 校验和
	checksum := calculateChecksum(encryptedData)

	// 创建消息对象
	msg := &Message{
		MsgID:     msgID,
		Type:      msgType,
		Data:      encryptedData,
		Timestamp: timestamp,
		Nonce:     nonce,
		Version:   "1.0",
		Checksum:  checksum,
	}

	// 生成签名
	msg.Sign = generateHMAC(msg, key)

	return msg, nil
}

// 发生消息
func SendMessage(conn net.Conn, msg *Message) error {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 将消息长度编码到前 4 个字节
	msgLen := make([]byte, MsgLenBytes)
	binary.BigEndian.PutUint32(msgLen, uint32(len(data)))

	// 拼接消息
	encodedMsg := append([]byte(MagicBytes), msgLen...)
	encodedMsg = append(encodedMsg, data...)

	fmt.Printf("Send:\n")
	fmt.Println(hex.EncodeToString(encodedMsg))
	fmt.Println()

	// 发送魔术字节、消息头和消息体
	_, err = conn.Write(encodedMsg)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// 发送消息到指定地址
func DialAndSend(host string, port int, msg *Message) error {
	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	return SendMessage(conn, msg)
}

// 从字节中获取消息并解密数据
func ParseMessage(data []byte, key string) ([]MessageInterface, error) {
	messages, err := decodeMessages(data)
	if err != nil {
		return nil, err
	}

	var validMessages []MessageInterface

	for _, msg := range messages {
		switch m := msg.(type) {
		case *Message:
			if err := verifyMessage(m, key); err != nil {
				return nil, err
			}
			decryptedData, err := encrypt.Decrypt(m.Data, key)
			if err != nil {
				return nil, err
			}
			m.Data = decryptedData
			// fmt.Printf("Received message: \n %+v \n", *msg)

			validMessages = append(validMessages, m)
		case *FileMessage:
			validMessages = append(validMessages, m)
		default:
			fmt.Println("Unknown message type")
		}
	}

	return validMessages, nil
}

// 从字节中解码消息
func decodeMessages(data []byte) ([]MessageInterface, error) {
	fmt.Printf("Recv:\n")
	fmt.Println(hex.EncodeToString(data))
	fmt.Println()

	var messages []MessageInterface
	buf := bytes.NewBuffer(data)

	for buf.Len() > 0 {
		// 读取并验证 Magic Bytes
		magicBytes := buf.Next(MagicBytesLen)
		var msgType = -1
		// 消息类型
		if bytes.Equal(magicBytes, []byte(MagicBytes)) {
			msgType = 0 // 普通消息
		} else if bytes.Equal(magicBytes, []byte(MagicBytes1)) {
			msgType = 1 // 文件消息
		}
		if msgType == -1 {
			return nil, errors.New("invalid magic bytes")
		}

		// 读取消息长度
		lengthBytes := buf.Next(MsgLenBytes)
		if len(lengthBytes) < MsgLenBytes {
			return nil, errors.New("invalid message length")
		}
		msgLen := binary.BigEndian.Uint32(lengthBytes)

		// 读取消息内容
		msgData := buf.Next(int(msgLen))
		if len(msgData) < int(msgLen) {
			return nil, ErrIncompleteMessage
		}

		// 反序列化消息内容
		switch msgType {
		// 业务消息
		case 0:
			var msg Message
			if err := json.Unmarshal(msgData, &msg); err != nil {
				return nil, err
			}
			messages = append(messages, &msg)
		// 文件消息
		case 1:
			var msg FileMessage
			if err := json.Unmarshal(msgData, &msg); err != nil {
				return nil, err
			}
			messages = append(messages, &msg)
		}
	}

	return messages, nil
}

// 校验消息
func verifyMessage(msg *Message, key string) error {
	// 验证时间戳是否过期
	if time.Since(time.Unix(msg.Timestamp, 0)) > time.Minute*5 {
		return errors.New("message is too old")
	}

	// 计算校验和并比较
	expectedChecksum := calculateChecksum(msg.Data)
	if msg.Checksum != expectedChecksum {
		return errors.New("checksum mismatch")
	}

	// 重新计算签名并比较
	sign := msg.Sign
	expectedSign := generateHMAC(msg, key)
	if sign != expectedSign {
		return errors.New("signature mismatch")
	}

	return nil
}

// 生成 HMAC 签名
func generateHMAC(msg *Message, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg.MsgID + msg.Data + msg.Nonce + msg.Version + msg.Checksum))
	return hex.EncodeToString(h.Sum(nil))
}

// 计算数据的校验和
func calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

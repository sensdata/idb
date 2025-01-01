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
	"github.com/sensdata/idb/core/utils"
)

// 消息起始字节
const (
	MagicBytes    = "\xAB\xCD\xEF\x01"
	MagicBytes1   = "\xAB\xCD\xEF\x02"
	MagicBytes2   = "\xAB\xCD\xEF\x03"
	MagicBytesLen = 4
	MsgLenBytes   = 4
)

// 消息类型
type MessageType string
type FileMessageType string
type SessionMessageType string

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
	FileErr  int = -1
	FileOk   int = 0
	FileDone int = 1
)

const (
	Start    SessionMessageType = "start"
	Detach   SessionMessageType = "detach"
	Attach   SessionMessageType = "attach"
	Finish   SessionMessageType = "finish"
	Rename   SessionMessageType = "rename"
	Transfer SessionMessageType = "transfer"
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

// Session 消息
type SessionMessage struct {
	MsgID     string             `json:"msg_id"`
	Type      SessionMessageType `json:"type"`
	Sign      string             `json:"sign"`
	Data      SessionData        `json:"data"`
	Timestamp int64              `json:"timestamp"`
	Nonce     string             `json:"nonce"`
	Version   string             `json:"version"`
	Checksum  string             `json:"checksum"`
}

type SessionData struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"`
}

func (m *SessionMessage) GetType() string {
	return "SessionMessage"
}

// ErrIncompleteMessage 表示接收到的消息数据不完整
var ErrIncompleteMessage = errors.New("incomplete message data")

func CreateFileMessage(msgID string, msgType FileMessageType, status int, path string, name string, total int64, offset int64, chunkSize int, chunk []byte) (*FileMessage, error) {
	// 检查 chunk 是否为空，允许 chunk 为空或小于 chunkSize
	if chunk == nil {
		// 如果 chunk 为空，确保 chunkSize 也设置为 0，以避免数据不一致
		chunkSize = 0
		chunk = []byte{} // 赋予一个空切片
	} else if len(chunk) < chunkSize {
		// 如果 chunk 长度不足 chunkSize，返回错误
		return nil, fmt.Errorf("chunk length is smaller than chunkSize: chunkSize=%d, chunk length=%d", chunkSize, len(chunk))
	}

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

	// fmt.Printf("Send:\n")
	// fmt.Println(hex.EncodeToString(encodedMsg))
	// fmt.Println()

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
	msg.Sign = generateHMAC(msg.MsgID, msg.Data, msg.Nonce, msg.Version, msg.Checksum, key)

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

	// fmt.Printf("Send:\n")
	// fmt.Println(hex.EncodeToString(encodedMsg))
	// fmt.Println()

	// 发送魔术字节、消息头和消息体
	_, err = conn.Write(encodedMsg)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

func CreateSessionMessage(msgID string, msgType SessionMessageType, data SessionData, key string, nonce string) (*SessionMessage, error) {
	// 时间戳
	timestamp := time.Now().Unix()

	// 结构
	dataJson, err := utils.ToJSONString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data")
	}

	// 校验和
	checksum := calculateChecksum(dataJson)

	// 创建消息
	msg := &SessionMessage{
		MsgID:     msgID,
		Type:      msgType,
		Data:      data,
		Timestamp: timestamp,
		Nonce:     nonce,
		Version:   "1.0",
		Checksum:  checksum,
	}

	// 生成签名
	msg.Sign = generateHMAC(msg.MsgID, dataJson, msg.Nonce, msg.Version, msg.Checksum, key)

	return msg, nil
}

// 发送消息
func SendSessionMessage(conn net.Conn, msg *SessionMessage) error {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 将消息长度编码到前 4 个字节
	msgLen := make([]byte, MsgLenBytes)
	binary.BigEndian.PutUint32(msgLen, uint32(len(data)))

	// 拼接消息
	encodedMsg := append([]byte(MagicBytes2), msgLen...)
	encodedMsg = append(encodedMsg, data...)

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

// 提取完整消息
func ExtractCompleteMessagePacket(buffer []byte) (int, []byte, []byte, error) {
	var msgType = -1

	// 至少需要 MagicBytesLen + MsgLenBytes 字节来解析消息头
	if len(buffer) < MagicBytesLen+MsgLenBytes {
		return msgType, nil, buffer, ErrIncompleteMessage
	}

	// 读取并验证 MagicBytes
	magicBytes := buffer[:MagicBytesLen]
	// 检查消息类型
	if bytes.Equal(magicBytes, []byte(MagicBytes)) {
		msgType = 0 // 普通消息
	} else if bytes.Equal(magicBytes, []byte(MagicBytes1)) {
		msgType = 1 // 文件消息
	} else if bytes.Equal(magicBytes, []byte(MagicBytes2)) {
		msgType = 2 // 会话消息
	} else {
		return msgType, nil, buffer, errors.New("invalid magic bytes")
	}

	// 读取消息长度
	lengthBytes := buffer[MagicBytesLen : MagicBytesLen+MsgLenBytes]
	msgLen := binary.BigEndian.Uint32(lengthBytes)

	// 计算完整消息的总长度（MagicBytes + MsgLenBytes + MsgData）
	totalMsgLen := MagicBytesLen + MsgLenBytes + int(msgLen)

	// 如果数据还不完整，返回 ErrIncompleteMessage
	if len(buffer) < totalMsgLen {
		return msgType, nil, buffer, ErrIncompleteMessage
	}

	// 提取完整消息
	msg := buffer[:totalMsgLen]

	// 返回剩余的数据缓冲区
	remainingBuffer := buffer[totalMsgLen:]

	return msgType, msg, remainingBuffer, nil
}

func DecodeMessage(msgType int, data []byte, key string) (MessageInterface, error) {
	// 反序列化消息内容
	switch msgType {
	// 业务消息
	case 0:
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}

		if err := verifyMessage(msg.Sign, msg.MsgID, msg.Data, msg.Timestamp, msg.Nonce, msg.Version, msg.Checksum, key); err != nil {
			return nil, err
		}
		decryptedData, err := encrypt.Decrypt(msg.Data, key)
		if err != nil {
			return nil, err
		}
		msg.Data = decryptedData
		return &msg, nil
	// 文件消息
	case 1:
		var msg FileMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	// 会话消息
	case 2:
		var msg SessionMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		dataJson, err := utils.ToJSONString(msg.Data)
		if err != nil {
			return nil, err
		}
		if err := verifyMessage(msg.Sign, msg.MsgID, dataJson, msg.Timestamp, msg.Nonce, msg.Version, msg.Checksum, key); err != nil {
			return nil, err
		}
		return &msg, nil
	default:
		return nil, errors.New("unsupported msg type")
	}
}

// 校验消息
func verifyMessage(sign string, msgID string, data string, timestamp int64, nonce string, version string, checksum string, key string) error {
	// 验证时间戳是否过期
	if time.Since(time.Unix(timestamp, 0)) > time.Minute*5 {
		return errors.New("message is too old")
	}

	// 计算校验和并比较
	expectedChecksum := calculateChecksum(data)
	if checksum != expectedChecksum {
		return errors.New("checksum mismatch")
	}

	// 重新计算签名并比较
	expectedSign := generateHMAC(msgID, data, nonce, version, checksum, key)
	if sign != expectedSign {
		return errors.New("signature mismatch")
	}

	return nil
}

//

// 生成 HMAC 签名
func generateHMAC(msgID string, data string, nonce string, version string, checksum string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msgID + data + nonce + version + checksum))
	return hex.EncodeToString(h.Sum(nil))
}

// 计算数据的校验和
func calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

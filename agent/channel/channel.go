package channel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sensdata/idb/agent/encrypt"
)

// 消息起始字节
const (
	MagicBytes    = "\xAB\xCD\xEF\x01"
	MagicBytesLen = 4
	MsgLenBytes   = 4
)

// 消息体
type Message struct {
	MsgID     string `json:"msg_id"`
	Sign      string `json:"sign"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Version   string `json:"version"`
	Checksum  string `json:"checksum"`
}

type ChannelService struct {
	key string
}

func NewChannelService(key string) *ChannelService {
	return &ChannelService{key: key}
}

func (s *ChannelService) AddMessage(data []byte) error {
	messages, err := s.decodeMessages(data)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if err := s.verifyMessage(msg); err != nil {
			return err
		}
		decryptedData, err := encrypt.Decrypt(msg.Data, s.key)
		if err != nil {
			return err
		}
		msg.Data = decryptedData
		log.Printf("Received message: %+v", msg)
	}

	return nil
}

func (s *ChannelService) decodeMessages(data []byte) ([]*Message, error) {
	var messages []*Message
	buf := bytes.NewBuffer(data)

	for buf.Len() > 0 {
		// 读取并验证 Magic Bytes
		magicBytes := buf.Next(MagicBytesLen)
		if !bytes.Equal(magicBytes, []byte(MagicBytes)) {
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
			return nil, errors.New("invalid message data")
		}

		// 反序列化消息内容
		var msg Message
		if err := json.Unmarshal(msgData, &msg); err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

// 验证接收到的消息
func (s *ChannelService) verifyMessage(msg *Message) error {
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
	msg.Sign = ""
	expectedSign := generateHMAC(msg, s.key)
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

// CreateMessage 创建并签名一个消息
func CreateMessage(msgID string, data string, key string, nonce string) (*Message, error) {
	timestamp := time.Now().Unix()
	checksum := calculateChecksum(data)

	// 加密数据
	encryptedData, err := encrypt.Encrypt(data, key)
	if err != nil {
		return nil, err
	}

	// 创建消息对象
	msg := &Message{
		MsgID:     msgID,
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

// 发送消息到指定地址
func SendMessage(host string, port int, msg *Message) error {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 将消息长度编码到前 4 个字节
	length := uint32(len(data))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, length)

	// 发送魔术字节、消息头和消息体
	_, err = conn.Write(append([]byte(MagicBytes), append(header, data...)...))
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// // 接收消息
// func ReceiveMessages(conn net.Conn, key string) ([]*Message, error) {
// 	var messages []*Message
// 	buf := make([]byte, 4096)
// 	tempBuf := new(bytes.Buffer)

// 	for {
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			if err == io.EOF {
// 				// 如果读取到文件结尾
// 				break
// 			}
// 			return nil, err
// 		}

// 		tempBuf.Write(buf[:n])

// 		for {
// 			// 确保我们至少有8个字节来读取魔术字节和消息头
// 			if tempBuf.Len() < 8 {
// 				break
// 			}

// 			// 寻找魔术字节
// 			magicIndex := bytes.Index(tempBuf.Bytes(), MagicBytes)
// 			if magicIndex == -1 {
// 				break
// 			}

// 			// 如果魔术字节之前有数据，丢弃这些数据
// 			if magicIndex > 0 {
// 				tempBuf.Next(magicIndex)
// 			}

// 			// 检查是否有完整的头部
// 			if tempBuf.Len() < 8 {
// 				break
// 			}

// 			// 读取消息头
// 			tempBuf.Next(4) // 跳过魔术字节
// 			header := tempBuf.Next(4)
// 			length := binary.BigEndian.Uint32(header)

// 			// 确保我们有足够的数据来读取完整的消息
// 			if uint32(tempBuf.Len()) < length {
// 				// 将消息头放回缓冲区，等待更多数据
// 				tempBuf.Write(MagicBytes)
// 				tempBuf.Write(header)
// 				break
// 			}

// 			// 读取完整的消息
// 			msgData := tempBuf.Next(int(length))
// 			var msg Message
// 			err := json.Unmarshal(msgData, &msg)
// 			if err != nil {
// 				return nil, err
// 			}

// 			// 验证消息
// 			err = VerifyMessage(&msg, key)
// 			if err != nil {
// 				return nil, err
// 			}

// 			// 解密数据
// 			decryptedData, err := encryption.Decrypt(msg.Data, []byte(key))
// 			if err != nil {
// 				return nil, err
// 			}
// 			msg.Data = string(decryptedData)

// 			// 将完整的消息添加到消息列表
// 			messages = append(messages, &msg)
// 		}
// 	}

// 	//返回已读取的消息
// 	return messages, nil
// }

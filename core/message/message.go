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
	MagicBytesLen = 4
	MsgLenBytes   = 4
)

// 消息类型
type MessageType string

const (
	Heartbeat     MessageType = "hb"
	CmdMessage    MessageType = "cmd"
	ActionMessage MessageType = "action"
)

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

// ErrIncompleteMessage 表示接收到的消息数据不完整
var ErrIncompleteMessage = errors.New("incomplete message data")

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
func ParseMessage(data []byte, key string) ([]*Message, error) {
	messages, err := decodeMessages(data)
	if err != nil {
		return nil, err
	}

	var validMessages []*Message

	for _, msg := range messages {
		if err := verifyMessage(msg, key); err != nil {
			return nil, err
		}
		decryptedData, err := encrypt.Decrypt(msg.Data, key)
		if err != nil {
			return nil, err
		}
		msg.Data = decryptedData
		// fmt.Printf("Received message: \n %+v \n", *msg)

		validMessages = append(validMessages, msg)
	}

	return validMessages, nil
}

// 从字节中解码消息
func decodeMessages(data []byte) ([]*Message, error) {
	fmt.Printf("Recv:\n")
	fmt.Println(hex.EncodeToString(data))
	fmt.Println()

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
			return nil, ErrIncompleteMessage
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

package utils

import (
	cryptoRand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

const (
	lettersAndDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	symbols          = "^&-_=+"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var randMu sync.Mutex

func GenerateMsgId() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateNonce(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GeneratePassword(length int) string {
	if length < 2 {
		// 至少需要2位：1个符号 + 1个其他字符
		return ""
	}

	randMu.Lock()
	defer randMu.Unlock()

	// 1. 生成一个符号字符
	symbol := symbols[seededRand.Intn(len(symbols))]

	// 2. 构造密码字符数组
	b := make([]byte, length)

	// 3. 确定符号出现的位置
	symbolPos := seededRand.Intn(length)
	b[symbolPos] = symbol

	// 4. 填充其他位置
	for i := range b {
		if i == symbolPos {
			continue
		}
		b[i] = lettersAndDigits[seededRand.Intn(len(lettersAndDigits))]
	}

	return string(b)
}

func GenerateUuid() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(cryptoRand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

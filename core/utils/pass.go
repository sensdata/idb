package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// HashPassword 使用 MD5 对密码和盐进行哈希
func HashPassword(password, salt string) string {
	saltedPassword := password + salt

	hasher := md5.New()
	hasher.Write([]byte(saltedPassword))
	hash := hasher.Sum(nil)

	return hex.EncodeToString(hash)
}

// ValidatePassword 验证密码
func ValidatePassword(storedHash, password, salt string) bool {
	inputHash := HashPassword(password, salt)
	return storedHash == inputHash
}

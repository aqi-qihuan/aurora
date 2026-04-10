package util

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// BCryptHash 生成BCrypt密码哈希（对标Java的BCryptPasswordEncoder）
func BCryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BCryptCheck 验证密码与哈希是否匹配
func BCryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5Hex 计算字符串的 MD5 哈希（十六进制）
// 用于: 文件唯一标识、API签名等
func MD5Hex(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256Hex 计算SHA-256哈希
func SHA256Hex(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Base64Encode Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64解码
func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// GenerateRandomString 生成指定长度的随机字符串
// 用于: 邀请码、重置令牌、临时密钥等
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// GenerateRandomStringSimple 生成随机字符串(忽略错误, 用于非安全场景)
func GenerateRandomStringSimple(length int) string {
	s, _ := GenerateRandomString(length)
	return s
}

// GenerateCode 生成数字验证码 (用于邮箱/手机验证码)
func GenerateCode(digits int) string {
	const numbers = "0123456789"
	b := make([]byte, digits)
	rand.Read(b) // 忽略错误
	for i := range b {
		b[i] = numbers[int(b[i])%len(numbers)]
	}
	return string(b)
}

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// GenerateTokenID 生成唯一的 Token ID
// 格式: {timestamp}_{random8chars}
func GenerateTokenID() (string, error) {
	ts := fmt.Sprintf("%d", time.Now().UnixNano())
	randStr, err := GenerateRandomString(8)
	if err != nil {
		return ts, nil
	}
	return ts + "_" + randStr, nil
}

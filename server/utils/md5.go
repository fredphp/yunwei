package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 MD5 加密
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// EncryptPassword 密码加密
func EncryptPassword(password, salt string) string {
	return MD5(MD5(password) + salt)
}

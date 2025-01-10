package iptv

import (
	"encoding/hex"
	"strings"

	"github.com/forgoer/openssl"
)

type TripleDESCrypto struct {
	key []byte
}

// NewTripleDESCrypto 创建新的3DES加密对象
func NewTripleDESCrypto(key string) *TripleDESCrypto {
	// 补齐密钥长度到24字节
	if len(key) < 24 {
		key += strings.Repeat("0", 24-len(key))
	} else if len(key) > 24 {
		key = key[:24]
	}

	return &TripleDESCrypto{
		key: []byte(key),
	}
}

// ECBEncrypt 加密函数，返回十六进制字符串
func (c *TripleDESCrypto) ECBEncrypt(plainText string) (string, error) {
	encrypted, err := openssl.Des3ECBEncrypt([]byte(plainText), c.key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encrypted), nil
}

// ECBDecrypt 解密函数，输入十六进制字符串，返回明文
func (c *TripleDESCrypto) ECBDecrypt(cipherText string) (string, error) {
	data, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	decrypted, err := openssl.Des3ECBDecrypt(data, c.key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

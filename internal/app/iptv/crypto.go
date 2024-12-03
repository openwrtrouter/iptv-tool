package iptv

import (
	"bytes"
	"crypto/des"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/forgoer/openssl"
)

// pad 补码函数，将文本填充至块大小的倍数（8字节）
func pad(src []byte) []byte {
	padding := des.BlockSize - len(src)%des.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// unPad 去补码函数，移除填充的数据
func unPad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("input is empty")
	}

	// 获取填充的长度
	unPadding := int(src[length-1])

	// 检查填充长度是否有效
	if unPadding <= 0 || unPadding > length {
		return nil, errors.New("invalid padding size")
	}

	return src[:(length - unPadding)], nil
}

// removePaddingCharacters 去除补码字符
func removePaddingCharacters(src []byte) ([]byte, error) {
	// 去除 '\x08' 字符
	src = bytes.ReplaceAll(src, []byte{'\x08'}, []byte{})
	return src, nil
}

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
	// 补码
	paddedText := pad([]byte(plainText))
	encrypted, err := openssl.Des3ECBEncrypt(paddedText, c.key, openssl.PKCS7_PADDING)
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

	// 去除补码
	result, err := removePaddingCharacters(decrypted)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

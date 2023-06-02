package cfgtable

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
)

const (
	Len16 = 16
	Len24 = 24
	Len32 = 32
)

func GenerateAesKey(skey []byte) []byte {
	has := md5.Sum(skey)
	return []byte(hex.EncodeToString(has[:]))
}

// AesCtrEncrypt 加密
func AesCtrEncrypt(plainText []byte, key []byte) ([]byte, error) {
	if len(key) != Len16 && len(key) != Len24 && len(key) != Len32 {
		return nil, ErrKeyLength
	}
	//1. 创建cipher.Block接口
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//2. 创建分组模式
	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)
	//3. 加密
	dst := make([]byte, len(plainText))
	stream.XORKeyStream(dst, plainText)

	return dst, nil
}

// AesCtrDecrypt 解密
func AesCtrDecrypt(encryptData []byte, key []byte) ([]byte, error) {
	return AesCtrEncrypt(encryptData, key)
}

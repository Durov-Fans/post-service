package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"strconv"
)

var (
	iv  = []byte("1234567890123456")
	key []byte
)

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func HashTgID(tgID int64) (string, error) {
	plain := pad([]byte(strconv.FormatInt(tgID, 10)))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plain))
	mode.CryptBlocks(encrypted, plain)

	return hex.EncodeToString(encrypted), nil
}

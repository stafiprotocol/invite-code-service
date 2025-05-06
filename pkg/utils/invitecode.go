package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 8

var charsetLen *big.Int

func init() {
	charsetLen = big.NewInt(int64(len(charset)))
}

func GenerateInviteCode() (string, error) {
	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}

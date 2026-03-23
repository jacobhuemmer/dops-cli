package crypto

import "strings"

type Encrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, "age1")
}

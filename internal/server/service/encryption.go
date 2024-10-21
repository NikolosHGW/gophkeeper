package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type EncryptionService struct {
	key []byte
}

func NewEncryptionService(key []byte) *EncryptionService {
	return &EncryptionService{key: key}
}

func (es *EncryptionService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("ошибка создания шифра: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (es *EncryptionService) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования base64: %w", err)
	}

	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("ошибка создания шифра: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("некорректный размер данных")
	}

	nonce, ciphertextData := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка расшифровки данных: %w", err)
	}

	return string(plaintext), nil
}

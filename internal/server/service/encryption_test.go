package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionService_EncryptDecrypt(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	es := NewEncryptionService(key)

	plaintext := "Тестовое сообщение для шифрования"

	ciphertext, err := es.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.NotEmpty(t, ciphertext)

	decryptedText, err := es.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decryptedText)
}

func TestEncryptionService_InvalidKey(t *testing.T) {
	key := []byte("короткий ключ")
	_, err := NewEncryptionService(key).Encrypt("текст")
	assert.Error(t, err)
}

func TestEncryptionService_DecryptWithWrongKey(t *testing.T) {
	key1 := []byte("01234567890123456789012345678901")
	key2 := []byte("10987654321098765432109876543210")

	es1 := NewEncryptionService(key1)
	es2 := NewEncryptionService(key2)

	plaintext := "Тестовое сообщение"

	ciphertext, err := es1.Encrypt(plaintext)
	assert.NoError(t, err)

	_, err = es2.Decrypt(ciphertext)
	assert.Error(t, err)
}

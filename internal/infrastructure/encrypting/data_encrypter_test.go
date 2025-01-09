package encrypting

import (
	"encoding/base64"
	"testing"

	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestAESEncrypter_EncryptDecryptJSON(t *testing.T) {
	cfg := &config.DataEncryptionConfig{
		Key:        "testkey1234567890",
		Salt:       "testsalt",
		Iterations: 10000,
		KeySize:    32,
	}

	encrypter := NewAESEncrypter(cfg)

	type TestData struct {
		Field1 string
		Field2 int
	}

	originalData := TestData{
		Field1: "test",
		Field2: 123,
	}

	encrypted, err := encrypter.EncryptJSON(&originalData)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	var decryptedData TestData
	err = encrypter.DecryptJSON(encrypted, &decryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

func TestAESEncrypter_EncryptJSON_Error(t *testing.T) {
	cfg := &config.DataEncryptionConfig{
		Key:        "testkey1234567890",
		Salt:       "testsalt",
		Iterations: 10000,
		KeySize:    32,
	}

	encrypter := NewAESEncrypter(cfg)

	// Test with data that cannot be marshaled to JSON
	_, err := encrypter.EncryptJSON(make(chan int))
	assert.Error(t, err)
}

func TestAESEncrypter_DecryptJSON_Error(t *testing.T) {
	cfg := &config.DataEncryptionConfig{
		Key:        "testkey1234567890",
		Salt:       "testsalt",
		Iterations: 10000,
		KeySize:    32,
	}

	encrypter := NewAESEncrypter(cfg)

	// Test with invalid base64 string
	err := encrypter.DecryptJSON("invalid_base64", &struct{}{})
	assert.Error(t, err)

	// Test with invalid ciphertext
	invalidCiphertext := base64.URLEncoding.EncodeToString([]byte("short"))
	err = encrypter.DecryptJSON(invalidCiphertext, &struct{}{})
	assert.Error(t, err)
}
package encrypting

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"golang.org/x/crypto/pbkdf2"
)

type Encrypter interface {
	EncryptJSON(data interface{}) (string, error)
	DecryptJSON(encrypted string, target interface{}) error
}

type AESEncrypter struct {
	key []byte
}

func NewAESEncrypter(cfg *config.DataEncryptionConfig) *AESEncrypter {
	salt := []byte(cfg.Salt)
	key := pbkdf2.Key([]byte(cfg.Key), salt, cfg.Iterations, cfg.KeySize, sha256.New)
	return &AESEncrypter{key: key}
}

func (e *AESEncrypter) EncryptJSON(data interface{}) (string, error) {
	// Marshal data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Create GCM cipher
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Encrypt and authenticate
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Combine nonce and ciphertext
	combined := make([]byte, len(nonce)+len(ciphertext))
	copy(combined, nonce)
	copy(combined[len(nonce):], ciphertext)

	// Encode to base64
	return base64.URLEncoding.EncodeToString(combined), nil
}

func (e *AESEncrypter) DecryptJSON(encrypted string, target interface{}) error {
	// Decode base64
	combined, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	// Split nonce and ciphertext
	if len(combined) < 12 {
		return fmt.Errorf("invalid ciphertext")
	}
	nonce := combined[:12]
	ciphertext := combined[12:]

	// Create cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return err
	}

	// Create GCM cipher
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Decrypt and verify
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decrypt error: %w", err)
	}

	// Unmarshal JSON
	return json.Unmarshal(plaintext, target)
}

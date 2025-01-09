package encrypting

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

// HasWithSalt generates a salt and hashes the password using PBKDF2.
// It returns the hashed password and the salt as hexadecimal strings.
func (h *Hasher) Hash(password string) (string, string, error) {
	// Generate a 16-byte random salt.
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	// Define PBKDF2 parameters.
	iterations := 100000
	keyLength := 32

	// Hash the password with the salt using PBKDF2.
	hash := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// Encode the hash and salt to hexadecimal strings.
	hashedPassword := hex.EncodeToString(hash)
	saltString := hex.EncodeToString(salt)

	return hashedPassword, saltString, nil
}

// Validate verifies if the provided password matches the hashed password using the salt.
// It returns true if the password is correct, false otherwise.
func (h *Hasher) Validate(password, hashedPassword, salt string) bool {
	// Decode the salt from hexadecimal to bytes.
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return false
	}

	// Define PBKDF2 parameters (must match those used during hashing).
	iterations := 100000
	keyLength := 32

	// Hash the input password with the salt.
	hash := pbkdf2.Key([]byte(password), saltBytes, iterations, keyLength, sha256.New)

	// Encode the computed hash to a hexadecimal string.
	computedHash := hex.EncodeToString(hash)

	// Compare the computed hash with the stored hashed password.
	return computedHash == hashedPassword
}

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// DefaultKey is a placeholder - should be overridden with a secure key
var DefaultKey = []byte("convoC2-default-key-change-me-now!!")

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts AES-GCM encrypted text
func Decrypt(encryptedBase64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aesGCM.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:aesGCM.NonceSize()], ciphertext[aesGCM.NonceSize():]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateKey generates a random AES key of specified length (16, 24, or 32 bytes)
func GenerateKey(length int) ([]byte, error) {
	if length != 16 && length != 24 && length != 32 {
		return nil, errors.New("key length must be 16, 24, or 32 bytes (128, 192, or 256 bits)")
	}

	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// Base64EncodeString encodes a string to base64
func Base64EncodeString(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64DecodeString decodes base64 to a string
func Base64DecodeString(encodedData string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

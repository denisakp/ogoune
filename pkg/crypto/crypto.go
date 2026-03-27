package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := loadKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key, err := loadKey()
	if err != nil {
		return "", err
	}

	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("ciphertext payload too short")
	}

	nonce, encrypted := payload[:nonceSize], payload[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func loadKey() ([]byte, error) {
	hexKey := os.Getenv("APP_SECRET_KEY")
	if hexKey == "" {
		return nil, errors.New("APP_SECRET_KEY is required")
	}
	if len(hexKey) != 64 {
		return nil, fmt.Errorf("APP_SECRET_KEY must be a 64-character hex string")
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("APP_SECRET_KEY must be valid hex: %w", err)
	}
	return key, nil
}

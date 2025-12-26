package cipher


import (
	"crypto/aes"
	"crypto/rand"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func GenerateMasterKey(keySize int) ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func EncryptData(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func DecryptData(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[0:gcm.NonceSize()]

	plaintext, err := gcm.Open(nil, nonce, ciphertext[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func GenerateHashID() (string, error) {
	id := make([]byte, 32)
	_, err := rand.Read(id)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(id)
	return hex.EncodeToString(hash[:]), nil
}

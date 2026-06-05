package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
)

type Crypto struct {
    key []byte
}

func NewCrypto(hexKey string) (*Crypto, error) {
    key, err := hex.DecodeString(hexKey)
    if err != nil || len(key) != 32 {
        return nil, errors.New("invalid 32-byte hex key")
    }
    return &Crypto{key: key}, nil
}

func (c *Crypto) KeyHash() []byte {
    h := sha256.Sum256(c.key)
    return h[:]
}

func (c *Crypto) VerifyKeyHash(hash []byte) bool {
    expected := c.KeyHash()
    if len(hash) != len(expected) {
        return false
    }
    for i := range hash {
        if hash[i] != expected[i] {
            return false
        }
    }
    return true
}

func (c *Crypto) Encrypt(plaintext []byte, nonce []byte) ([]byte, error) {
    block, err := aes.NewCipher(c.key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    if len(nonce) != gcm.NonceSize() {
        return nil, errors.New("nonce size must match GCM nonce size")
    }
    ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
    return ciphertext, nil
}

func (c *Crypto) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
    block, err := aes.NewCipher(c.key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    return gcm.Open(nil, nonce, ciphertext, nil)
}

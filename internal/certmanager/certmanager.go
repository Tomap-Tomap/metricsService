// Package certmanager stores structures for working with message encryption
package certmanager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
)

// EncryptManager message encryption structure
type EncryptManager struct {
	publicKey *rsa.PublicKey
}

// NewEncryptManager create EncryptManager
func NewEncryptManager(pathToFile string) (*EncryptManager, error) {
	publicKeyPEM, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}

	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &EncryptManager{publicKey.(*rsa.PublicKey)}, nil
}

// EncryptMessage returns an encrypted message
func (em *EncryptManager) EncryptMessage(m []byte) ([]byte, error) {
	encryptData, err := rsa.EncryptPKCS1v15(rand.Reader, em.publicKey, m)
	if err != nil {
		return nil, fmt.Errorf("encrypt message: %w", err)
	}

	return encryptData, nil
}

// DecryptManager message decryption structure
type DecryptManager struct {
	privateKey *rsa.PrivateKey
}

// NewDecryptManager create DecryptManager
func NewDecryptManager(pathToFile string) (*DecryptManager, error) {
	privateKeyPEM, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &DecryptManager{privateKey.(*rsa.PrivateKey)}, nil
}

// DecryptMessage decrypts the message
func (dm *DecryptManager) DecryptMessage(m []byte) ([]byte, error) {
	decryptData, err := rsa.DecryptPKCS1v15(rand.Reader, dm.privateKey, m)
	if err != nil {
		return nil, fmt.Errorf("decrypt message: %w", err)
	}

	return decryptData, nil
}

// RequestDecrypt middleware to decrypt the message
func (dm *DecryptManager) RequestDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		decrypt, err := dm.DecryptMessage(buf.Bytes())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf.Reset()
		buf.Write(decrypt)

		r.Body = io.NopCloser(&buf)
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"pingspot/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	if _, err := os.Stat("keys"); os.IsNotExist(err) {
		if err := os.Mkdir("keys", 0755); err != nil {
			logger.Error("Failed to create keys directory: ", zap.Error(err))
			return
		}
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	privateFile, err := os.Create("keys/private.pem")
	if err != nil {
		logger.Error("Failed to create private key file: ", zap.Error(err))
		return
	}
	defer privateFile.Close()

	pem.Encode(privateFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKey := &privateKey.PublicKey
	publicBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	publicFile, err := os.Create("keys/public.pem")
	if err != nil {
		logger.Error("Failed to create public key file: ", zap.Error(err))
		return
	}
	defer publicFile.Close()

	pem.Encode(publicFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})

	logger.Info("Keys successfully generated..")
}

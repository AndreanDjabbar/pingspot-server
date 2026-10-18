package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"pingspot/internal/logger"

	"go.uber.org/zap"
)

func main() {
	if _, err := os.Stat("certs"); os.IsNotExist(err) {
		if err := os.Mkdir("certs", 0755); err != nil {
			logger.Error("Failed to create certs directory: ", zap.Error(err))
			return
		}
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	privateFile, err := os.Create("certs/private.pem")
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
	publicFile, err := os.Create("certs/public.pem")
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

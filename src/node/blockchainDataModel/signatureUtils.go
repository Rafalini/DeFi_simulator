package blockchainDataModel

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func ReturnHash(message []byte) []byte {
	hash := sha256.New()
	hash.Write(message)
	return hash.Sum(nil)
}

func GenerateKeyPairAndReturn(prefix string) (*rsa.PrivateKey, rsa.PublicKey) {
	// Generate a new RSA key pair with a key size of 2048 bits
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	// Encode the private key to ASN.1 DER format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	// Create a PEM block for the private key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write the private key to a file
	privateKeyFile, _ := os.Create(prefix + "_private_key.pem")
	defer privateKeyFile.Close()
	pem.Encode(privateKeyFile, privateKeyPEM)

	// Extract the public key from the private key
	publicKey := privateKey.PublicKey

	// Marshal the public key to ASN.1 DER format
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&publicKey)

	// Create a PEM block for the public key
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write the public key to a file
	publicKeyFile, _ := os.Create(prefix + "_public_key.pem")
	defer publicKeyFile.Close()
	pem.Encode(publicKeyFile, publicKeyPEM)

	return privateKey, publicKey
}

func ReadKeyFile(filename string) ([]byte, error) {
	// Read the contents of the file
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return keyData, nil
}

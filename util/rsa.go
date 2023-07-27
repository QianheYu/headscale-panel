package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var caCertPool = x509.NewCertPool()

// GetCAPool returns the certificate pool for the Certificate Authority
func GetCAPool() *x509.CertPool {
	return caCertPool
}

// LoadCA appends the CA certificate content to the certificate pool
func LoadCA(ca []byte) bool {
	return caCertPool.AppendCertsFromPEM(ca)
}

// LoadCAFromFile loads the CA certificate from a file and appends it to the certificate pool
func LoadCAFromFile(file string) error {
	caCert, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Cannot open ca cert")
	}
	if !LoadCA(caCert) {
		return fmt.Errorf("Not add ca cert")
	}
	return nil
}

// EncodeStr2Base64 encrypting base64 strings
func EncodeStr2Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// DecodeStrFromBase64 decrypting base64 strings
func DecodeStrFromBase64(str string) string {
	decodeBytes, _ := base64.StdEncoding.DecodeString(str)
	return string(decodeBytes)
}

// LoadPrivateKey loads a private key from a file and returns it
func LoadPrivateKey(fileName, password string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	var privKey any
	switch block.Type {
	case "ENCRYPTED PRIVATE KEY":
		if x509.IsEncryptedPEMBlock(block) {
			// Warn: Since we won't be using blocks here,
			// we've chosen to reuse block.Bytes,
			// so be sure to change this if you use blocks in subsequent extensions.
			block.Bytes, err = x509.DecryptPEMBlock(block, []byte(password))
			if err != nil {
				return nil, fmt.Errorf("DecryptPEMBlock error: %v", err)
			}
		}
		fallthrough
	case "PRIVATE KEY":
		privKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("Unknown or Unsupported private key type")
	}

	if err != nil {
		return nil, err
	}

	switch privKey.(type) {
	case *rsa.PrivateKey:
		return privKey.(*rsa.PrivateKey), nil
	// case *ecdsa.PrivateKey:

	default:
		return nil, fmt.Errorf("Unsupported private key type")
	}
}

// LoadPublicKey loads a public key from a file and returns it
func LoadPublicKey(fileName string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}

// RSADecrypt decrypts the given base64 data using the private key
func RSADecrypt(base64Data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	data := []byte(DecodeStrFromBase64(string(base64Data)))
	res, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return res, fmt.Errorf("Cannot decrypt, private key may be incorrect, %v", err)
	}
	return res, nil
}

// RSAEncrypt encrypts the given data using the public key and returns it
func RSAEncrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	res, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return res, fmt.Errorf("Cannot encrypt, public key may be incorrect, %v", err)
	}
	return []byte(EncodeStr2Base64(string(res))), nil
}

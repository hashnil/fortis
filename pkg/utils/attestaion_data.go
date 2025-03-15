package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fortis/infrastructure/config"
	"log"
)

// GetAttestationData is a simulator utility for WebAuthN credentials to be used by backend
func GetAttestationData(challenge string) map[string]interface{} {
	clientData, _ := json.Marshal(map[string]string{
		"challenge": challenge,
		"type":      "key.create",
	})

	block, _ := pem.Decode(config.DFNS_PRIVATE_KEY)
	if block == nil {
		log.Fatalf("failed to decode PEM block")
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("failed to parse RSA private key: %v", err)
	}

	// Type assert the key to *rsa.PrivateKey
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("parsed key is not an RSA private key")
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("error marshaling public key: %v", err)
	}

	publicKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}))

	h := sha256.Sum256(clientData)
	credInfo, _ := json.Marshal(map[string]string{
		"clientDataHash": hex.EncodeToString(h[:]),
		"publicKey":      publicKeyPEM,
	})

	s := sha256.Sum256(credInfo)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, s[:])
	if err != nil {
		log.Fatalf("error signing data: %v", err)
	}

	attestationData, _ := json.Marshal(map[string]string{
		"publicKey": publicKeyPEM,
		"signature": hex.EncodeToString(signature),
	})

	requestPayload := map[string]interface{}{
		"firstFactorCredential": map[string]interface{}{
			"credentialKind": "Key",
			"credentialInfo": map[string]string{
				"credId":          config.GetCredentialID(),
				"clientData":      toBase64URL(clientData),
				"attestationData": toBase64URL(attestationData),
			},
		},
	}

	return requestPayload
}

func toBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

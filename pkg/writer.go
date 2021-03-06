package gsw

import (
	"context"
	"encoding/base64"

	"github.com/google/go-github/v31/github"
)

// copy from https://github.com/doodlesbykumbi/github-secrets-writer

// NewSecretWriter creates a SecretWriter that uses the Github OAuth token provided for API calls.
func NewSecretWriter(client *github.Client) *SecretWriter {
	return &SecretWriter{
		client: client,
	}
}

// SecretWriter provides the ability to create and update Github secrets.
type SecretWriter struct {
	client *github.Client
}

// Write encrypts and writes a Github secret to Github using the API.
func (s SecretWriter) Write(owner, repo, secretName string, secretValue []byte) (string, error) {
	publicKeyId, publicKey, err := s.getPublicKey(owner, repo)
	if err != nil {
		return "", err
	}

	encryptedValue, err := encryptValue(secretValue, publicKey)
	if err != nil {
		return "", err
	}

	res, err := s.client.Actions.CreateOrUpdateSecret(
		context.Background(),
		owner,
		repo,
		&github.EncryptedSecret{
			Name:           secretName,
			KeyID:          publicKeyId,
			EncryptedValue: base64.StdEncoding.EncodeToString(encryptedValue),
		})
	if err != nil {
		return "", err
	}

	return res.Status, nil
}

func (s SecretWriter) getPublicKey(
	owner, repo string,
) (string, *[publicKeyLength]byte, error) {
	pk, _, err := s.client.Actions.GetPublicKey(
		context.Background(),
		owner,
		repo)
	if err != nil {
		return "", nil, err
	}

	keyId := pk.GetKeyID()
	key64String := pk.GetKey()

	keySlice, err := base64.StdEncoding.DecodeString(key64String)

	var publicKey = &[32]byte{}
	copy(publicKey[:], keySlice)
	if err != nil {
		return "", nil, err
	}

	return keyId, publicKey, nil
}

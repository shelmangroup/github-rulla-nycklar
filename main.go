package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gsw "github.com/shelmangroup/github-secrets-sync/pkg"
	"google.golang.org/api/iam/v1"
)

var owner = "shelmangroup"
var repo = "github-secrets-sync"
var secretName = "TEST"
var secretValue = []byte("super secret value")
var testEmail = "github-test@xXxXx.iam.gserviceaccount.com"
var testProject = "xXxXx"

type IamServiceAccountClient struct {
	service *iam.Service
}

func NewIamClient() *IamServiceAccountClient {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		log.Fatalf("iam.NewService: %v", err)
	}
	return &IamServiceAccountClient{service: service}
}

// createKey creates a service account key.
func (i *IamServiceAccountClient) createKey(serviceAccountEmail string) (*iam.ServiceAccountKey, error) {
	resource := "projects/-/serviceAccounts/" + serviceAccountEmail
	request := &iam.CreateServiceAccountKeyRequest{}
	key, err := i.service.Projects.ServiceAccounts.Keys.Create(resource, request).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.Create: %v", err)
	}
	log.Printf("Created key: %v", key.Name)
	return key, nil
}

// deleteKey deletes a service account key.
func (i *IamServiceAccountClient) deleteKey(fullKeyName string) error {
	var err error
	_, err = i.service.Projects.ServiceAccounts.Keys.Delete(fullKeyName).Do()
	if err != nil {
		return fmt.Errorf("Projects.ServiceAccounts.Keys.Delete: %v", err)
	}
	log.Printf("Deleted key: %v", fullKeyName)
	return nil
}

// listKey lists a service account's keys.
func (i *IamServiceAccountClient) listKeys(serviceAccountEmail string) ([]*iam.ServiceAccountKey, error) {
	resource := "projects/-/serviceAccounts/" + serviceAccountEmail
	response, err := i.service.Projects.ServiceAccounts.Keys.List(resource).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.List: %v", err)
	}
	for _, key := range response.Keys {
		log.Printf("Listing key: %v", key.Name)
	}
	return response.Keys, nil
}

// rotateKey, and remove old keys if exists
func (i *IamServiceAccountClient) rotateKey(serviceAccountEmail string) (*iam.ServiceAccountKey, error) {
	keys, err := i.listKeys(serviceAccountEmail)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		err = i.deleteKey(key.Name)
		if err != nil {
			return nil, err
		}
	}

	key, err := i.createKey(serviceAccountEmail)
	if err != nil {
		return nil, err
	}
	return key, nil

}

func main() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}

	writer := gsw.NewSecretWriter(token)
	status, err := writer.Write(owner, repo, secretName, secretValue)
	if err != nil {
		log.Printf("Ops.. %s\n", err.Error())
	} else {
		log.Printf("secret write status: %s\n", status)
	}

}

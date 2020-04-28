package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	gsw "github.com/shelmangroup/github-secrets-sync/pkg"
	"google.golang.org/api/iam/v1"
)

var owner = "shelmangroup"
var repo = "github-secrets-sync"
var secretName = "TEST"
var secretValue = []byte("super secret value")

// createKey creates a service account key.
func rotateServiceAccountKey(w io.Writer, serviceAccountEmail string) (*iam.ServiceAccountKey, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService: %v", err)
	}

	resource := "projects/-/serviceAccounts/" + serviceAccountEmail
	request := &iam.CreateServiceAccountKeyRequest{}
	key, err := service.Projects.ServiceAccounts.Keys.Create(resource, request).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.Create: %v", err)
	}
	fmt.Fprintf(w, "Created key: %v", key.Name)
	return key, nil
}

func main() {
	fmt.Println("vim-go")

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

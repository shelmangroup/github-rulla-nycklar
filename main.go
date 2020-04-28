package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v31/github"
	gsw "github.com/shelmangroup/github-secrets-sync/pkg"
	"golang.org/x/oauth2"
)

var client *github.Client
var ctx = context.Background()
var owner = "shelmangroup"
var repo = "github-secrets-sync"
var secretName = "TEST"
var secretValue = []byte("super secret value")

func main() {
	fmt.Println("vim-go")

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	secret, _, err := client.Actions.GetSecret(ctx, owner, repo, secretName)
	if err != nil {
		log.Printf("Ops.. %s\n", err.Error())
	} else {
		log.Printf("secret: %+v", secret)
	}

	writer := gsw.NewSecretWriter(token)
	status, err := writer.Write(owner, repo, secretName, secretValue)
	if err != nil {
		log.Printf("Ops.. %s\n", err.Error())
	} else {
		log.Printf("secret write status: %s\n", status)
	}

}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

var client *github.Client
var ctx = context.Background()

func main() {
	fmt.Println("vim-go")

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	secret, _, err := client.Actions.GetSecret(ctx, "mad01", "dummy", "FOO")
	if err != nil {
		log.Printf("Ops.. %s\n", err.Error())
	} else {
		log.Printf("secret: %+v", secret)
	}
}

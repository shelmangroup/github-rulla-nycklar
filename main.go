package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v31/github"

	"github.com/bradleyfalzon/ghinstallation"
	joonix "github.com/joonix/log"
	gsw "github.com/shelmangroup/github-secrets-sync/pkg"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iam/v1"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logJSON    = kingpin.Flag("log-json", "Use structured logging in JSON format").Default("false").Bool()
	logFluentd = kingpin.Flag("log-fluentd", "Use structured logging in GKE Fluentd format").Default("false").Bool()
	logLevel   = kingpin.Flag("log-level", "The level of logging").Default("info").Enum("debug", "info", "warn", "error", "panic", "fatal")
	keyFile    = kingpin.Flag("github-key-file", "PEM file for signed requests").Required().ExistingFile()
	appID      = kingpin.Flag("github-app-id", "GitHub App ID").Required().Int64()

	owner = kingpin.Flag("owner", "Github Owner/User").Required().String()
	repo  = kingpin.Flag("repo", "Github Repo").Required().String()

	// "github-test@xXxXx.iam.gserviceaccount.com"
	serviceAccountEmail = kingpin.Flag("service-account", "Google Service Account Email").Required().String()

	secretName = kingpin.Flag("secret-name", "Github Secret name").Default("GOOGLE_APPLICATION_CREDENTIALS").String()
)

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

	// cant be deleted, should always exist even if no user keys have ben added
	systemManagedKey := keys[len(keys)-1]

	for _, key := range keys {
		if systemManagedKey.Name == key.Name {
			continue
		}
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

	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.DefaultEnvars()
	kingpin.Parse()

	switch strings.ToLower(*logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if *logJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if *logFluentd {
		log.SetFormatter(joonix.NewFormatter())
	}

	log.SetOutput(os.Stderr)

	/*	// rotate and get new key
		iamClient := NewIamClient()
		key, err := iamClient.rotateKey(*serviceAccountEmail)
		if err != nil {
			log.Fatal(err)
		}
		keyDecoded, _ := base64.URLEncoding.DecodeString(key.PrivateKeyData)
		log.Println(string(keyDecoded))
		//
	*/

	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport
	atr, err := ghinstallation.NewAppsTransportKeyFromFile(tr, *appID, *keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	githubClient := github.NewClient(&http.Client{Transport: atr})
	//

	//
	ctx := context.Background()
	installs, _, err := githubClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	//

	for _, install := range installs {
		log.Debugf("installationID: %v", install.GetID())
		log.Debugf("Install: %+v", install)

		itr := ghinstallation.NewFromAppsTransport(atr, install.GetID())
		ic := github.NewClient(&http.Client{Transport: itr})

		writer := gsw.NewSecretWriter(ic)
		status, err := writer.Write(*owner, *repo, *secretName, []byte("super secret string"))
		if err != nil {
			log.Printf("Ops.. %s\n", err.Error())
		} else {
			log.Printf("secret write status: %s\n", status)
		}

	}

}

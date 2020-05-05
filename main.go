package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/google/go-github/v31/github"

	"github.com/bradleyfalzon/ghinstallation"
	joonix "github.com/joonix/log"
	gsw "github.com/shelmangroup/github-rulla-nycklar/pkg"
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
	installid  = kingpin.Flag("github-install-id", "GitHub Install ID").Required().Int64()

	owner                   = kingpin.Flag("owner", "Github Owner/User").Required().String()
	repoToServiceAccountMap = kingpin.Flag("repo-to-email", "Google service account to github repo in format of repo=email").Required().StringMap()

	// "github-rulla-nycklar=github-test@xXxXx.iam.gserviceaccount.com"

	secretName = kingpin.Flag("secret-name", "Github Secret name").Default("GOOGLE_APPLICATION_CREDENTIALS").String()
)

func validateRepoToServiceAccountMap(input map[string]string) bool {
	knownEmails := make(map[string]string)
	for repo, email := range input {
		log.Debugf("validate repo email input repo=email (%v=%v)", repo, email)
		ok := validateGoogleServiceAccountEmail(email)
		if !ok {
			return false
		}

		if _, present := knownEmails[email]; present {
			log.Errorf("service account email (%v) is already used, duplicate service accounts is not supported", email)
			return false
		}

		knownEmails[email] = ""
	}

	return true
}

func validateGoogleServiceAccountEmail(email string) bool {
	match, _ := regexp.MatchString("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.iam\\.gserviceaccount\\.com$)", email)
	return match
}

type IamServiceAccountClient struct {
	service *iam.Service
	ctx     *context.Context
}

func NewIamClient(ctx context.Context) *IamServiceAccountClient {
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
	log.Debugf(spew.Sprintf("service account keys: %+v", keys))

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

	ok := validateRepoToServiceAccountMap(*repoToServiceAccountMap)
	if !ok {
		log.Fatalf("invalid input from flag --repo-to-email got: %v", *repoToServiceAccountMap)
	}

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

	// Shared transport to reuse TCP connections.
	transport := http.DefaultTransport
	appTransport, err := ghinstallation.NewAppsTransportKeyFromFile(transport, *appID, *keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	installTransport := ghinstallation.NewFromAppsTransport(appTransport, *installid)
	githubClient := github.NewClient(&http.Client{Transport: installTransport})
	secretWriter := gsw.NewSecretWriter(githubClient)
	//

	//
	ctx := context.Background()
	iamClient := NewIamClient(ctx)
	//

	/*
		To find the install id on github go to
		Org > Settings > Installed Github Apps > AppName > Configure
		in the URL you can see the install ID
		https://github.com/organizations/<ORG>/settings/installations/<install id>
	*/

	getKey := func(email string) []byte {
		key, err := iamClient.rotateKey(email)
		if err != nil {
			log.Fatal(err)
		}
		keyDecoded, _ := base64.URLEncoding.DecodeString(key.PrivateKeyData)
		return keyDecoded
	}
	//

	writeSecret := func(repo string, key []byte) {
		status, err := secretWriter.Write(*owner, repo, *secretName, key)
		if err != nil {
			log.Errorf("Ops.. %s\n", err.Error())
		} else {
			log.Infof("secret write status: %s\n", status)
		}
	}

	for repo, email := range *repoToServiceAccountMap {
		log.Debugf("repo=email (%v=%v)", repo, email)
		keyBytes := getKey(email)
		writeSecret(repo, keyBytes)
	}

}

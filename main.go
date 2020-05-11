package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v31/github"

	"github.com/bradleyfalzon/ghinstallation"
	joonix "github.com/joonix/log"
	gsw "github.com/shelmangroup/github-rulla-nycklar/pkg"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
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

func main() {

	var (
		logJSON        = kingpin.Flag("log-json", "Use structured logging in JSON format").Default("false").Bool()
		logFluentd     = kingpin.Flag("log-fluentd", "Use structured logging in GKE Fluentd format").Default("false").Bool()
		logLevel       = kingpin.Flag("log-level", "The level of logging").Default("info").Enum("debug", "info", "warn", "error", "panic", "fatal")
		keyFile        = kingpin.Flag("github-key-file", "PEM file for signed requests").Required().ExistingFile()
		appID          = kingpin.Flag("github-app-id", "GitHub App ID").Required().Int64()
		installid      = kingpin.Flag("github-install-id", "GitHub Install ID").Required().Int64()
		owner          = kingpin.Flag("owner", "Github Owner/User").Required().String()
		repoToEmailMap = kingpin.Flag("repo-to-email", "Google service account to github repo in format of repo=email").Required().StringMap()
		secretName     = kingpin.Flag("secret-name", "Github Secret name").Default("GOOGLE_APPLICATION_CREDENTIALS").String()
	)

	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.DefaultEnvars()
	kingpin.Parse()

	ok := validateRepoToServiceAccountMap(*repoToEmailMap)
	if !ok {
		log.Fatalf("invalid input from flag --repo-to-email got: %v", *repoToEmailMap)
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

	run(*appID, *installid, *owner, *keyFile, *secretName, *repoToEmailMap)

}

func run(appID, installID int64, owner, privateKeyFile, secretName string, repoToEmail map[string]string) {
	// Shared transport to reuse TCP connections.
	transport := http.DefaultTransport
	appTransport, err := ghinstallation.NewAppsTransportKeyFromFile(transport, appID, privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	installTransport := ghinstallation.NewFromAppsTransport(appTransport, installID)
	githubClient := github.NewClient(&http.Client{Transport: installTransport})
	secretWriter := gsw.NewSecretWriter(githubClient)
	//

	//
	ctx := context.Background()
	iamClient := NewIamClient(ctx)
	//

	getKey := func(email string) []byte {
		key, err := iamClient.rotateKey(email)
		if err != nil {
			log.Fatal(err)
		}
		keyDecoded, _ := base64.URLEncoding.DecodeString(key.PrivateKeyData)
		return keyDecoded
	}
	//

	writeSecret := func(repo string, key []byte) error {
		status, err := secretWriter.Write(owner, repo, secretName, key)
		if err != nil {
			log.Errorf("Ops.. %s\n", err.Error())
			return err
		} else {
			log.Infof("secret write status: %s\n", status)
		}
		return nil
	}

	for repo, email := range repoToEmail {
		log.Debugf("repo=email (%v=%v)", repo, email)
		keyBytes := getKey(email)
		err := writeSecret(repo, keyBytes)
		if err != nil {
			log.Fatalf("Ops.. %s\n", err.Error())
		}
	}
}

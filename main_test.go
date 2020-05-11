package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServiceAccount(t *testing.T) {
	testCases := []struct {
		expected bool
		email    string
	}{

		{
			expected: true,
			email:    "test@one-123.iam.gserviceaccount.com",
		},

		{
			expected: true,
			email:    "test@two-prod-234.iam.gserviceaccount.com",
		},

		{
			expected: true,
			email:    "test@o-234.iam.gserviceaccount.com",
		},

		{
			expected: false,
			email:    "test@two-prod-234-iam.gserviceaccount.com",
		},

		{
			expected: false,
			email:    "test@example.com",
		},

		{
			expected: false,
			email:    "test@asdf.iam.example.com",
		},

		{
			expected: false,
			email:    "asdifasdf.com",
		},

		{
			expected: false,
			email:    "api.example..com",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, validateGoogleServiceAccountEmail(tc.email), tc.email)
	}
}

func TestValidateRepoToServiceAccountMap(t *testing.T) {
	testCases := []struct {
		expected bool
		input    map[string]string
	}{
		{
			expected: true,
			input: map[string]string{
				"foo": "test-foo@one-123.iam.gserviceaccount.com",
				"bar": "test-bar@two-123.iam.gserviceaccount.com",
			},
		},

		// the same service account is used in multiple repos
		{
			expected: false,
			input: map[string]string{
				"foo": "test-foo@one-123.iam.gserviceaccount.com",
				"bar": "test-foo@one-123.iam.gserviceaccount.com",
			},
		},

		{
			expected: false,
			input: map[string]string{
				"foo": "foo@example.com",
				"bar": "bar@example.com",
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, validateRepoToServiceAccountMap(tc.input), tc.input)
	}
}

func TestCLI(t *testing.T) {
	t.Skip("only for use when debugging locally to do live debugging")
	var (
		appID          int64 = 62880
		installID      int64 = 8467539
		owner                = "shelmangroup"
		privateKeyFile       = "rulla-nyckel-bot.2020-04-29.private-key.pem"
		secretName           = "SuperHemligSecretGoland"
	)

	repoToEmail := map[string]string{
		"test-foo": "github-test-foo@dev-123.iam.gserviceaccount.com",
		"test-bar": "github-test-bar@dev-1600.iam.gserviceaccount.com",
	}

	run(appID, installID, owner, privateKeyFile, secretName, repoToEmail)
}

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
		// "github-test@xXxXx.iam.gserviceaccount.com"

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

func TestProjectFromServiceAccount(t *testing.T) {
	testCases := []struct {
		expected string
		email    string
	}{
		// "github-test@xXxXx.iam.gserviceaccount.com"

		{
			expected: "one-123",
			email:    "test@one-123.iam.gserviceaccount.com",
		},

		{
			expected: "two-prod-123",
			email:    "test@two-prod-123.iam.gserviceaccount.com",
		},

		{
			expected: "",
			email:    "test@two-prod-234-iam.gserviceaccount.com",
		},

		{
			expected: "",
			email:    "two-prod-234-iam.gserviceaccount.com",
		},

		{
			expected: "",
			email:    "two-prod-234-iam.example.com",
		},

		{
			expected: "",
			email:    "foo@two-prod-234-iam.example.com",
		},

		{
			expected: "",
			email:    "foo@.example.com",
		},

		{
			expected: "",
			email:    "example.com",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, getProjectFromServiceAccount(tc.email), tc.email)
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

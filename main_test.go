package main

import (
	"fmt"
	"testing"

	"google.golang.org/api/iam/v1"

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

func TestIsSystemManagedKey(t *testing.T) {
	testCases := []struct {
		expected bool
		input    *iam.ServiceAccountKey
	}{
		{
			expected: false,
			input: &iam.ServiceAccountKey{
				ValidAfterTime:  "2020-05-05T13:34:26Z",
				ValidBeforeTime: "9999-12-31T23:59:59Z",
			},
		},

		{
			expected: true,
			input: &iam.ServiceAccountKey{
				ValidAfterTime:  "2020-05-05T13:34:26Z",
				ValidBeforeTime: "2022-05-08T04:58:36Z",
			},
		},
	}

	iamClient := IamServiceAccountClient{}
	for _, tc := range testCases {
		msg := fmt.Sprintf("ValidAfterTime: (%v) ValidBeforeTime: (%v)", tc.input.ValidAfterTime, tc.input.ValidBeforeTime)
		assert.Equal(t, tc.expected, iamClient.isSystemMangedKey(tc.input), msg)
	}
}

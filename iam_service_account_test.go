package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iam/v1"
)

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

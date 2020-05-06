package main

import (
	"fmt"
	"reflect"
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

func TestKeysToDelete(t *testing.T) {
	testCases := []struct {
		expected []*iam.ServiceAccountKey
		input    []*iam.ServiceAccountKey
	}{
		{
			expected: []*iam.ServiceAccountKey{
				{ValidAfterTime: "2020-05-05T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "b"},
			},
			input: []*iam.ServiceAccountKey{
				{ValidAfterTime: "2020-05-07T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "d"},
				{ValidAfterTime: "2020-05-06T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "c"},
				{ValidAfterTime: "2020-05-04T13:34:26Z", ValidBeforeTime: "2022-05-08T04:58:36Z", Name: "a"},
				{ValidAfterTime: "2020-05-05T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "b"},
			},
		},

		{
			expected: []*iam.ServiceAccountKey{},
			input: []*iam.ServiceAccountKey{
				{ValidAfterTime: "2020-05-06T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "c"},
				{ValidAfterTime: "2020-05-04T13:34:26Z", ValidBeforeTime: "2022-05-08T04:58:36Z", Name: "a"},
				{ValidAfterTime: "2020-05-05T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "b"},
			},
		},

		{
			expected: []*iam.ServiceAccountKey{},
			input: []*iam.ServiceAccountKey{
				{ValidAfterTime: "2020-05-04T13:34:26Z", ValidBeforeTime: "2022-05-08T04:58:36Z", Name: "a"},
				{ValidAfterTime: "2020-05-05T13:34:26Z", ValidBeforeTime: "9999-12-31T23:59:59Z", Name: "b"},
			},
		},
	}

	iamClient := IamServiceAccountClient{}
	for _, tc := range testCases {
		output := iamClient.keysToDelete(tc.input)
		assert.Equal(t, len(tc.expected), len(output))
		if len(tc.expected) == 0 && len(output) == 0 {
			continue
		}
		assert.True(t, reflect.DeepEqual(output, tc.expected), tc.input)

	}
}

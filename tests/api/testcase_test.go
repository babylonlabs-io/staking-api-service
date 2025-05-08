//go:build e2e

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testcase struct {
	testName         string
	endpoint         string
	expectedHttpCode int
	expectedContents string
}

func checkCases(t *testing.T, cases []testcase) {
	for _, cs := range cases {
		t.Run(cs.testName, func(t *testing.T) {
			t.Parallel()
			assertResponse(t, cs.endpoint, cs.expectedHttpCode, cs.expectedContents)
		})
	}
}

func assertResponse(t *testing.T, endpoint string, expectedHttpCode int, expectedJSON string) {
	t.Helper()

	contents, code := clientGet(t, endpoint)
	assert.Equal(t, expectedHttpCode, code)
	assert.JSONEqf(t, expectedJSON, string(contents), "received json: %s", contents)
}

/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workflow

import (
	"testing"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRaiseErrorFailureDetails(t *testing.T) {
	t.Run("parses structured failure details", func(t *testing.T) {
		details, err := parseRaiseErrorFailureDetails(`{
			"errorType": "PaymentFailed",
			"errorMessage": "Payment was declined",
			"stackTrace": "PaymentService.Process",
			"innerFailure": {
				"errorType": "ProviderError",
				"isNonRetriable": true
			},
			"isNonRetriable": true
		}`)
		require.NoError(t, err)
		assert.Equal(t, "PaymentFailed", details.GetErrorType())
		assert.Equal(t, "Payment was declined", details.GetErrorMessage())
		assert.Equal(t, "PaymentService.Process", details.GetStackTrace().GetValue())
		require.NotNil(t, details.GetInnerFailure())
		assert.Equal(t, "ProviderError", details.GetInnerFailure().GetErrorType())
		assert.True(t, details.GetInnerFailure().GetIsNonRetriable())
		assert.True(t, details.GetIsNonRetriable())
	})

	t.Run("rejects empty details", func(t *testing.T) {
		_, err := parseRaiseErrorFailureDetails("  ")
		require.EqualError(t, err, "--failure-details is required")
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		_, err := parseRaiseErrorFailureDetails(`{"errorType":`)
		require.ErrorContains(t, err, "invalid --failure-details JSON")
	})

	t.Run("requires a non-empty error type", func(t *testing.T) {
		_, err := parseRaiseErrorFailureDetails(`{"errorMessage":"failed"}`)
		require.EqualError(t, err, "--failure-details must include a non-empty errorType")
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		_, err := parseRaiseErrorFailureDetails(`{"errorType":"Failure","unknown":true}`)
		require.ErrorContains(t, err, "invalid --failure-details JSON")
	})
}

func TestRaiseErrorFailureDetailsMatchesDurableTaskContract(t *testing.T) {
	details, err := parseRaiseErrorFailureDetails(`{"errorType":"Failure","errorMessage":"failed"}`)
	require.NoError(t, err)
	assert.Equal(t, &protos.TaskFailureDetails{
		ErrorType:    "Failure",
		ErrorMessage: "failed",
	}, details)
}

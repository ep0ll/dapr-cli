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
	"errors"
	"fmt"
	"strings"

	"github.com/dapr/cli/pkg/print"
	"github.com/dapr/cli/pkg/workflow"
	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/kit/signals"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

var flagRaiseErrorFailureDetails string

var RaiseErrorCmd = &cobra.Command{
	Use:   "raise-error <instance-id>/<event-name>",
	Short: "Raise an error for a workflow waiting for an external event.",
	Long: "Raise an error for a workflow waiting for an external event. " +
		"The --failure-details flag must contain a JSON object with a non-empty errorType. " +
		"For example: --failure-details='{"errorType":"PaymentFailed","errorMessage":"Payment was declined"}'.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("the argument must be in the format '<instance-id>/<event-name>'")
		}

		failureDetails, err := parseRaiseErrorFailureDetails(flagRaiseErrorFailureDetails)
		if err != nil {
			return err
		}

		appID, err := getWorkflowAppID(cmd)
		if err != nil {
			return err
		}

		err = workflow.RaiseError(signals.Context(), workflow.RaiseErrorOptions{
			KubernetesMode: flagKubernetesMode,
			Namespace:      flagDaprNamespace,
			AppID:          appID,
			InstanceID:     parts[0],
			Name:           parts[1],
			FailureDetails: failureDetails,
		})
		if err != nil {
			print.FailureStatusEvent(cmd.ErrOrStderr(), err.Error())
			return err
		}

		print.InfoStatusEvent(cmd.OutOrStdout(), "Workflow '%s' raised error for event '%s' successfully", parts[0], parts[1])
		return nil
	},
}

func parseRaiseErrorFailureDetails(raw string) (*protos.TaskFailureDetails, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("--failure-details is required")
	}

	details := new(protos.TaskFailureDetails)
	if err := protojson.Unmarshal([]byte(raw), details); err != nil {
		return nil, fmt.Errorf("invalid --failure-details JSON: %w", err)
	}
	if strings.TrimSpace(details.GetErrorType()) == "" {
		return nil, errors.New("--failure-details must include a non-empty errorType")
	}

	return details, nil
}

func init() {
	RaiseErrorCmd.Flags().StringVarP(
		&flagRaiseErrorFailureDetails,
		"failure-details",
		"d",
		"",
		"Structured workflow failure details as a JSON object. The errorType field is required.",
	)
	_ = RaiseErrorCmd.MarkFlagRequired("failure-details")

	WorkflowCmd.AddCommand(RaiseErrorCmd)
}

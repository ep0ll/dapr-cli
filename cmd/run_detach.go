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

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dapr/cli/pkg/print"
)

const detachedRunEnv = "DAPR_CLI_DETACHED_RUN"

func startDetachedRun(argv []string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve dapr executable: %w", err)
	}

	args, err := removeDetachFlag(argv)
	if err != nil {
		return err
	}

	logPath, logFile, err := createDetachedRunLog()
	if err != nil {
		return err
	}
	defer logFile.Close()

	child := exec.Command(executable, args...)
	child.Env = append(os.Environ(), detachedRunEnv+"=1")
	child.Stdin = nil
	child.Stdout = logFile
	child.Stderr = logFile
	if err := configureDetachedProcess(child); err != nil {
		return fmt.Errorf("configure detached process: %w", err)
	}
	if err := child.Start(); err != nil {
		return fmt.Errorf("start detached Dapr run: %w", err)
	}
	// The child owns its process lifetime and log file descriptor from this point.
	if err := child.Process.Release(); err != nil {
		return fmt.Errorf("release detached Dapr process: %w", err)
	}

	print.SuccessStatusEvent(os.Stdout, "Dapr run started in detached mode. Logs: %s", logPath)
	return nil
}

func removeDetachFlag(argv []string) ([]string, error) {
	result := make([]string, 0, len(argv))
	flags := true
	for _, arg := range argv {
		if flags && arg == "--" {
			flags = false
			result = append(result, arg)
			continue
		}
		if flags && (arg == "--detach" || arg == "-d") {
			continue
		}
		if flags && strings.HasPrefix(arg, "--detach=") {
			value := strings.TrimPrefix(arg, "--detach=")
			if value != "true" && value != "false" {
				return nil, fmt.Errorf("invalid value for --detach: %q", value)
			}
			continue
		}
		result = append(result, arg)
	}
	return result, nil
}

func createDetachedRunLog() (string, *os.File, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	logDir := filepath.Join(configDir, "dapr", "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return "", nil, fmt.Errorf("create detached run log directory: %w", err)
	}

	app := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, appID)
	if app == "" {
		app = "dapr"
	}
	name := fmt.Sprintf("%s-%s.log", app, time.Now().Format("20060102-150405.000000000"))
	path := filepath.Join(logDir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", nil, fmt.Errorf("create detached run log file: %w", err)
	}
	return path, file, nil
}

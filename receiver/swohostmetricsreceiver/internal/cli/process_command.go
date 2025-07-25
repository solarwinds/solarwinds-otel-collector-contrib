// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ProcessJSONCommand executes the command that outputs JSON data within the CommandLineExecutor and returns that data unmarshalled
// Returns an error instead of data when command or unmarshalling fails (also logs this error with additional info packed inside).
func ProcessJSONCommand[TResult any](
	cle CommandLineExecutor,
	command string,
	logger *zap.Logger,
) (TResult, error) {
	// actual scraping
	stdout, stderr, err := cle.ExecuteCommand(command)
	if err != nil || stderr != "" {
		logExecutionError(command, stdout, stderr, err, logger)
		if err == nil {
			err = fmt.Errorf("%s", stderr)
		}
		return *new(TResult), err
	}

	var result TResult
	err = json.Unmarshal([]byte(stdout), &result)
	if err != nil {
		zap.L().Error(
			fmt.Sprintf("Unmarshaling result of command %s failed", command),
			zap.Error(err),
		)
		return *new(TResult), err
	}

	zap.L().Debug(fmt.Sprintf("Command %s succeeded with result %+v", command, result))
	return result, nil
}

// ProcessCommand executes passed command with the help of CommandLineExecutor.
// Returns output from stdout and error indicator.
func ProcessCommand(cle CommandLineExecutor, command string, logger *zap.Logger) (string, error) {
	stdout, stderr, err := cle.ExecuteCommand(command)
	if err != nil || stderr != "" {
		logExecutionError(command, stdout, stderr, err, logger)
		if err == nil {
			err = fmt.Errorf("%s", stderr)
		}
		return stdout, err
	}

	logger.Debug(fmt.Sprintf("Command %s succeeded with result %+v", command, stdout))
	return stdout, nil
}

// Logs command, stdout, stderr and error with the Error level.
func logExecutionError(command, stdout, stderr string, err error, logger *zap.Logger) {
	logger.Error(
		fmt.Sprintf(
			"Command %s failed.",
			command,
		),
		zap.String("command", command),
		zap.String("stdout", stdout),
		zap.String("stderr", stderr),
		zap.Error(err),
	)
}

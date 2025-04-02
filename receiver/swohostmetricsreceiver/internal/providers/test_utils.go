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

package providers

import (
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/cli"
)

type cliExecutorMock struct {
	stdout string
	stderr string
	err    error
}

var _ cli.CommandLineExecutor = (*cliExecutorMock)(nil)

func CreateCommandLineExecutorMock(
	stdout string,
	stderr string,
	err error,
) cli.CommandLineExecutor {
	return &cliExecutorMock{
		stdout: stdout,
		stderr: stderr,
		err:    err,
	}
}

// ExecuteCommand implements cli.CommandLineExecutor.
func (e *cliExecutorMock) ExecuteCommand(_ ...string) (stdOut string, stdErr string, error error) {
	return e.stdout, e.stderr, e.err
}

type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
}

type cliMultiCommandExecutorMock struct {
	commandSet map[string]CommandResult
}

var _ cli.CommandLineExecutor = (*cliMultiCommandExecutorMock)(nil)

func CreateMultiCommandExecutorMock(commandsWithResults map[string]CommandResult) cli.CommandLineExecutor {
	return &cliMultiCommandExecutorMock{
		commandSet: commandsWithResults,
	}
}

// ExecuteCommand implements cli.CommandLineExecutor.
func (m *cliMultiCommandExecutorMock) ExecuteCommand(command ...string) (stdout string, stderr string, error error) {
	combinedCommand := strings.Join(command, " ")
	if v, found := m.commandSet[combinedCommand]; found {
		return v.Stdout, v.Stderr, v.Err
	}
	return "", "", nil
}

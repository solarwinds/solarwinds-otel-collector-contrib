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
	"fmt"
)

type BashCli struct{}

var _ (CommandLineExecutor) = (*BashCli)(nil)

func NewBashCliExecutor() CommandLineExecutor {
	return new(BashCli)
}

func (*BashCli) ExecuteCommand(command ...string) (stdOut, stdErr string, error error) {
	internalCommands := append([]string{"-c"}, command...)
	stdOut, stdErr, err := executeCommand(
		"bash",
		internalCommands...)
	if err != nil {
		message := fmt.Sprintf("Bash command %s failed.", command)
		return "", "", fmt.Errorf(message, err)
	}

	return stdOut, stdErr, nil
}

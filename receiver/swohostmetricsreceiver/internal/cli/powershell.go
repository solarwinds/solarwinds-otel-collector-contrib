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

import "fmt"

type PowershellCli struct{}

var _ (CommandLineExecutor) = (*PowershellCli)(nil)

func (*PowershellCli) ExecuteCommand(command ...string) (stdOut, strErr string, error error) {
	stdOut, stdErr, err := executeCommand(
		"powershell",
		command...,
	)
	if err != nil {
		message := fmt.Sprintf("Powershell command %s failed.", command)
		return "", "", fmt.Errorf(message, err)
	}

	return stdOut, stdErr, nil
}

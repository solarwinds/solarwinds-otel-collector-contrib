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

package loggedusers

import (
	"fmt"
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/cli"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

const (
	// Command last reads from wtmp file by default. Returns users
	// with active session. For parsing purposes are last two lines omitted.
	lastCommand = "last -R | head -n -2"
)

type provider struct {
	cli cli.CommandLineExecutor
}

var _ providers.Provider[Data] = (*provider)(nil)

func CreateProvider() providers.Provider[Data] {
	return &provider{
		cli: &cli.BashCli{},
	}
}

// Provide implements Provider.
func (lup *provider) Provide() <-chan Data {
	ch := make(chan Data)
	go func() {
		defer close(ch)
		stdout, err := cli.ProcessCommand(lup.cli, lastCommand)
		result := Data{
			Error: err,
		}
		if err == nil {
			result.Users = getUsers(stdout)
		}
		zap.L().Debug(fmt.Sprintf("LoggedUsers provider result: %+v", result))
		ch <- result
	}()
	return ch
}

// getUsers goes through lastCommand output and parses
// users with active session.
func getUsers(output string) []User {
	var loggedInUsers []User
	loginRecords := strings.Split(output, "\n")

	for _, record := range loginRecords {
		// used to omit system boot session with suffix "still running"
		// and other not active records
		record = strings.TrimSpace(record)
		if !strings.HasSuffix(record, "still logged in") {
			continue
		}
		// username is in first column, separated by space
		// Linux distribution should not contain spaces in the username
		username := parseNthColumn(record, 0)
		tty := parseNthColumn(record, 1)

		if username != "" {
			loggedInUsers = append(loggedInUsers, User{
				Name: username,
				TTY:  tty,
			})
		}
	}
	return loggedInUsers
}

// parseNthColumn takes text and index and returns nth column. If the index
// is out of the range, the parseNthColumn returns empty string.
func parseNthColumn(text string, index int) string {
	splitText := strings.Fields(text)
	if len(splitText) <= index {
		return ""
	}
	return splitText[index]
}

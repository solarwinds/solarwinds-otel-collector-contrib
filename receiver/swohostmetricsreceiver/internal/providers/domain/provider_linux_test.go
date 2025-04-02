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

package domain

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/stretchr/testify/assert"
)

func Test_Provide_ProvidesCompleteData(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "swdev.local",
		FQDN:       "swdev.local.SWI-D5ZRKQ2",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "swdev.local.SWI-D5ZRKQ2",
			Stderr: "",
			Err:    nil,
		},
		domainCommand: {
			Stdout: "swdev.local",
			Stderr: "",
			Err:    nil,
		},
	}

	processActWithEvaluation(t, commands, expectedDomain)
}

func Test_Provide_ProvidesPartialDomainOnFqdnFailure(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "swdev.local",
		FQDN:       "",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "",
			Stderr: "",
			Err:    fmt.Errorf("hostname is not available"),
		},
		domainCommand: {
			Stdout: "swdev.local",
			Stderr: "",
			Err:    nil,
		},
	}
	processActWithEvaluation(t, commands, expectedDomain)
}

func Test_Provide_ProvidesPartialDomainOnDomainFailure(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "",
		FQDN:       "swdev.local.SWI-D5ZRKQ2",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "swdev.local.SWI-D5ZRKQ2",
			Stderr: "",
			Err:    nil,
		},
		domainCommand: {
			Stdout: "swdev.local",
			Stderr: "",
			Err:    fmt.Errorf("hostname is not available"),
		},
	}
	processActWithEvaluation(t, commands, expectedDomain)
}

func Test_Provide_ProvidesPartialDomainOnDomainStderrFailure(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "",
		FQDN:       "swdev.local.SWI-D5ZRKQ2",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "swdev.local.SWI-D5ZRKQ2",
			Stderr: "",
			Err:    nil,
		},
		domainCommand: {
			Stdout: "",
			Stderr: "some error",
			Err:    nil,
		},
	}
	processActWithEvaluation(t, commands, expectedDomain)
}

func Test_Provide_ProvidesPartialDomainOnFqdnStderrFailure(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "swdev.local",
		FQDN:       "",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "",
			Stderr: "some error",
			Err:    nil,
		},
		domainCommand: {
			Stdout: "swdev.local",
			Stderr: "",
			Err:    nil,
		},
	}
	processActWithEvaluation(t, commands, expectedDomain)
}

func Test_Provide_ProvidesInvalidObjectOnBothFailures(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "",
		FQDN:       "",
		DomainRole: -1,
		Workgroup:  "",
	}

	commands := map[string]providers.CommandResult{
		fqdnCommand: {
			Stdout: "",
			Stderr: "",
			Err:    fmt.Errorf("hostname is not available"),
		},
		domainCommand: {
			Stdout: "swdev.local",
			Stderr: "",
			Err:    fmt.Errorf("hostname is not available"),
		},
	}

	processActWithEvaluation(t, commands, expectedDomain)
}

func processActWithEvaluation(
	t *testing.T,
	commands map[string]providers.CommandResult,
	expectedDomain Domain,
) {
	sut := provider{
		cli: providers.CreateMultiCommandExecutorMock(commands),
	}

	ch := sut.Provide()
	actualDomain := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedDomain, actualDomain)
	assert.False(t, open, "channel must be closed")
}

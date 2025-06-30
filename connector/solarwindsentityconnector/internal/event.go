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

package internal

import (
	"fmt"
	"go.opentelemetry.io/collector/pdata/plog"
)

// Event interface defines methods for handling events related to entities and relationships.
type Event interface {
	Update(logRecords *plog.LogRecordSlice)
	Delete(logRecords *plog.LogRecordSlice)
}

func GetActionType(e Event) (string, error) {
	switch e.(type) {
	case Entity:
		return GetActionString(e.(Entity).Action)
	case *Relationship:
		return GetActionString(e.(*Relationship).Action)
	default:
		return "", fmt.Errorf("unsupported event type: %T", e)
	}
}

func GetActionString(input string) (string, error) {
	if input == EventUpdateAction {
		return EventUpdateAction, nil
	} else if input == EventDeleteAction {
		return EventDeleteAction, nil
	}
	return "", fmt.Errorf("failed to get action type from input: %s", input)
}

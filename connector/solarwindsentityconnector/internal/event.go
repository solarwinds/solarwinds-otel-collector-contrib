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
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

type EventBuilder struct{}

func NewEventBuilder() *EventBuilder {
	return &EventBuilder{}
}

func (e *EventBuilder) AppendUpdateEvent(eventLogs *plog.LogRecordSlice, subject Subject) {
	lr := eventLogs.AppendEmpty()
	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	subject.Update(&lr)
}

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

package mqttreceiver

import "time"

// MQTT configuration constants
const (
	// QoSLevel defines the MQTT Quality of Service level
	QoSLevel = 1

	// PingTopic is the topic used for health check pings
	PingTopic = "ping"

	// PingInterval defines how often to send ping messages
	PingInterval = 10 * time.Second
)

// Status constants for broker connection states
const (
	StatusOK               = "OK"
	StatusConnectionFailed = "ConnectionFailed"
	StatusRoundtripFailed  = "RoundtripFailed"
	StatusSubscribeFailed  = "SubscribeFailed"
)

// BrokerMetrics contains all predefined metric definitions

const (
	TopicClientsConnected         = "$SYS/broker/clients/connected"
	TopicClientSubscriptionsCount = "$SYS/broker/subscriptions/count"
	TopicBytesReceived            = "$SYS/broker/load/bytes/received/1min"
	TopicBytesSent                = "$SYS/broker/load/bytes/sent/1min"
	TopicMessagesReceived         = "$SYS/broker/load/messages/received/1min"
	TopicMessagesSent             = "$SYS/broker/load/messages/sent/1min"
)

var buildInBrokerTopicList = []string{
	TopicClientsConnected,
	TopicClientSubscriptionsCount,
	TopicBytesReceived,
	TopicBytesSent,
	TopicMessagesReceived,
	TopicMessagesSent,
}

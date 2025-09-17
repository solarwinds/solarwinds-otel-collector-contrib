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

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/mqttreceiver/internal/metadata"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type subscriptionMetadata struct {
	broker *Broker
	sensor *Sensor
	metric *Metric
}

func newSubscriptionMetadata(broker *Broker, sensor *Sensor, metric *Metric) *subscriptionMetadata {
	return &subscriptionMetadata{
		broker: broker,
		sensor: sensor,
		metric: metric,
	}
}

// brokerSubscription encapsulates MQTT broker connection and subscription management
type brokerSubscription struct {
	broker                 *Broker
	client                 *mqtt.Client
	logger                 *zap.Logger
	consumer               consumer.Metrics
	ctx                    context.Context
	wg                     sync.WaitGroup
	topicSubscriptions     map[string][]*subscriptionMetadata
	settings               receiver.Settings
	roundtripMetricBuilder *metadata.MetricsBuilder
}

// newBrokerSubscription creates a new subscribed broker instance
func newBrokerSubscription(broker *Broker, settings receiver.Settings, logger *zap.Logger, consumer consumer.Metrics, ctx context.Context) *brokerSubscription {
	topicSubscriptions := make(map[string][]*subscriptionMetadata)

	for _, sensor := range broker.Sensors {
		for _, metric := range sensor.Metrics {
			if _, exists := topicSubscriptions[metric.Topic]; !exists {
				topicSubscriptions[metric.Topic] = []*subscriptionMetadata{}
			}
			topicSubscriptions[metric.Topic] = append(topicSubscriptions[metric.Topic], newSubscriptionMetadata(broker, sensor, metric))
		}
	}

	for _, metric := range brokerMetrics {
		if _, exists := topicSubscriptions[metric.Topic]; !exists {
			topicSubscriptions[metric.Topic] = []*subscriptionMetadata{}
		}
		topicSubscriptions[metric.Topic] = append(topicSubscriptions[metric.Topic], newSubscriptionMetadata(broker, nil, metric))
	}

	roundtripMetricBuilder := metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), settings)

	return &brokerSubscription{
		broker:                 broker,
		logger:                 logger,
		consumer:               consumer,
		ctx:                    ctx,
		topicSubscriptions:     topicSubscriptions,
		settings:               settings,
		roundtripMetricBuilder: roundtripMetricBuilder,
	}
}

// Start connects to the broker and starts all subscriptions
func (sb *brokerSubscription) Start() error {
	client, err := sb.createMQTTClient()
	if err != nil {
		return fmt.Errorf("failed to create MQTT client: %w", err)
	}
	sb.client = client

	if err := sb.connectClient(); err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}

	if err := sb.subscribeTopics(); err != nil {
		sb.logger.Error("Failed to subscribe to sensor metrics",
			zap.String("broker", sb.broker.Name),
			zap.Error(err))
	}

	sb.startHealthCheck()

	return nil
}

// Stop disconnects from the broker and cleans up resources
func (sb *brokerSubscription) Stop() {
	if sb.client != nil && (*sb.client).IsConnected() {
		(*sb.client).Disconnect(250)
	}
	sb.wg.Wait()
}

// IsConnected returns whether the broker is currently connected
func (sb *brokerSubscription) IsConnected() bool {
	return sb.client != nil && (*sb.client).IsConnected()
}

func (sb *brokerSubscription) createMQTTClient() (*mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%d", sb.broker.Protocol, sb.broker.Server, sb.broker.Port))
	opts.SetClientID(fmt.Sprintf("otel-mqtt-receiver-%d", time.Now().Unix()))
	opts.SetUsername(sb.broker.User)
	opts.SetPassword(sb.broker.Password)
	opts.SetDefaultPublishHandler(sb.defaultMessageHandler)

	// Add connection lost handler
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		sb.logger.Error("MQTT connection lost",
			zap.String("broker", sb.broker.Name),
			zap.Error(err))
		sb.sendBrokerStatus(StatusConnectionFailed)
	})

	client := mqtt.NewClient(opts)
	return &client, nil
}

func (sb *brokerSubscription) connectClient() error {
	if token := (*sb.client).Connect(); token.Wait() && token.Error() != nil {
		sb.logger.Error("Failed to connect to broker",
			zap.String("broker", sb.broker.Name),
			zap.Error(token.Error()))
		sb.sendBrokerStatus(StatusConnectionFailed)
		return token.Error()
	}

	sb.logger.Info("Connected to MQTT broker",
		zap.String("broker", sb.broker.Name),
		zap.String("server", sb.broker.Server),
		zap.Int("port", sb.broker.Port))

	return nil
}

func (sb *brokerSubscription) subscribeTopics() error {
	topics := make(map[string]byte, len(sb.topicSubscriptions))

	for topic := range sb.topicSubscriptions {
		topics[topic] = QoSLevel
	}

	if token := (*sb.client).SubscribeMultiple(topics, sb.defaultMessageHandler); token.Wait() && token.Error() != nil {
		sb.logger.Error("Failed to subscribe to topics",
			zap.Error(token.Error()))
		return token.Error()
	}

	sb.logger.Info("Subscribed to topics",
		zap.Int("count", len(topics)),
		zap.String("broker", sb.broker.Name))

	return nil
}

func (sb *brokerSubscription) defaultMessageHandler(_ mqtt.Client, msg mqtt.Message) {
	metricMetadata, exists := sb.topicSubscriptions[msg.Topic()]
	if !exists {
		sb.logger.Warn("Received message on unknown topic",
			zap.String("topic", msg.Topic()),
			zap.String("broker", sb.broker.Name))
		return
	}
	for _, mMetadata := range metricMetadata {
		err := sb.handleMessage(msg, mMetadata)
		if err != nil {
			sb.logger.Error("Failed to handle message",
				zap.String("topic", msg.Topic()),
				zap.String("broker", sb.broker.Name),
				zap.Error(err))
		}
	}
}

// handleMessage processes incoming MQTT messages and converts them to OpenTelemetry metrics
func (sb *brokerSubscription) handleMessage(message mqtt.Message, metadata *subscriptionMetadata) error {
	value, err := extractValue(message, metadata.metric.JsonProperty)
	if err != nil {
		return err
	}

	metrics, err := sb.createMetrics(metadata, &value)
	if err != nil {
		return err
	}

	return sb.consumer.ConsumeMetrics(sb.ctx, metrics)
}

func (sb *brokerSubscription) startHealthCheck() {
	sb.wg.Add(1)
	go sb.runHealthCheck()
}

func (sb *brokerSubscription) runHealthCheck() {
	defer sb.wg.Done()

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	var startTime time.Time
	var roundtripCompleted = true

	// Subscribe to ping response
	pingHandler := func(client mqtt.Client, msg mqtt.Message) {
		roundtripCompleted = true
		duration := time.Since(startTime)

		sb.logger.Info("Roundtrip completed",
			zap.String("broker", sb.broker.Name),
			zap.Duration("duration", duration))

		if err := sb.sendRoundtripMetric(duration); err != nil {
			sb.logger.Error("Failed to send roundtrip metric",
				zap.String("broker", sb.broker.Name),
				zap.Error(err))
		}
	}

	if token := (*sb.client).Subscribe(PingTopic, QoSLevel, pingHandler); token.Wait() && token.Error() != nil {
		sb.logger.Error("Failed to subscribe to ping topic",
			zap.String("broker", sb.broker.Name),
			zap.Error(token.Error()))
		return
	}

	sb.logger.Debug("Started health check", zap.String("broker", sb.broker.Name))

	for {
		select {
		case <-sb.ctx.Done():
			return
		case <-ticker.C:
			if err := sb.performHealthCheck(&startTime, &roundtripCompleted, pingHandler); err != nil {
				sb.logger.Error("Health check failed",
					zap.String("broker", sb.broker.Name),
					zap.Error(err))
			}
		}
	}
}

func (sb *brokerSubscription) performHealthCheck(startTime *time.Time, roundtripCompleted *bool, pingHandler mqtt.MessageHandler) error {
	if !*roundtripCompleted {
		sb.logger.Error("Roundtrip timeout", zap.String("broker", sb.broker.Name))
		sb.sendBrokerStatus(StatusRoundtripFailed)
	}

	if !(*sb.client).IsConnected() {
		if err := sb.reconnectAndResubscribe(pingHandler); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	*startTime = time.Now()
	*roundtripCompleted = false

	if token := (*sb.client).Publish(PingTopic, QoSLevel, false, []byte("ping")); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish ping: %w", token.Error())
	}

	return nil
}

func (sb *brokerSubscription) reconnectAndResubscribe(pingHandler mqtt.MessageHandler) error {
	sb.logger.Info("Attempting to reconnect", zap.String("broker", sb.broker.Name))

	if token := (*sb.client).Connect(); token.Wait() && token.Error() != nil {
		sb.sendBrokerStatus(StatusConnectionFailed)
		return token.Error()
	}

	if token := (*sb.client).Subscribe(PingTopic, QoSLevel, pingHandler); token.Wait() && token.Error() != nil {
		sb.sendBrokerStatus(StatusSubscribeFailed)
		return token.Error()
	}

	sb.logger.Info("Reconnected successfully", zap.String("broker", sb.broker.Name))
	return nil
}

// createRoundtripMetric creates a roundtrip metric using generated MetricsBuilder
func (sb *brokerSubscription) createRoundtripMetric(duration time.Duration) pmetric.Metrics {
	sb.roundtripMetricBuilder.RecordSwOtelcolMqttRoundtripDataPoint(
		pcommon.NewTimestampFromTime(time.Now()), float64(duration.Milliseconds()))
	rb := metadata.NewResourceBuilder(metadata.DefaultResourceAttributesConfig())
	setBrokerAttributes(rb, sb.broker)
	rb.SetSwOtelcolMqttBrokerStatusOK()
	res := rb.Emit()
	return sb.roundtripMetricBuilder.Emit(metadata.WithResource(res))
}

// sendRoundtripMetric creates and sends a roundtrip metric
func (sb *brokerSubscription) sendRoundtripMetric(duration time.Duration) error {
	return sb.consumer.ConsumeMetrics(sb.ctx, sb.createRoundtripMetric(duration))
}

// sendBrokerStatus sends a broker status metric
func (sb *brokerSubscription) sendBrokerStatus(status string) {
	rb := metadata.NewResourceBuilder(metadata.DefaultResourceAttributesConfig())
	setBrokerAttributes(rb, sb.broker)
	setStatus(rb, status)
	res := rb.Emit()
	metrics := sb.newMetricsWithResource(res)
	rm := metrics.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	m := sm.Metrics().AppendEmpty()
	m.SetName("sw.otelcol.MqttBroker.Status")
	m.SetDescription("Broker status")
	m.SetEmptyGauge()
	dp := m.Gauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetIntValue(statusToInt(status))

	if err := sb.consumer.ConsumeMetrics(sb.ctx, metrics); err != nil {
		sb.logger.Error("Failed to send broker status",
			zap.String("broker", sb.broker.Name),
			zap.String("status", status),
			zap.Error(err))
	}
}

// statusToInt maps status strings to integer values
func statusToInt(status string) int64 {
	switch status {
	case StatusOK:
		return 0
	case StatusConnectionFailed:
		return 1
	case StatusRoundtripFailed:
		return 2
	case StatusSubscribeFailed:
		return 3
	default:
		return -1
	}
}

// newMetricsWithResource returns a pmetric.Metrics initialized with the provided resource and instrumentation scope.
func (sb *brokerSubscription) newMetricsWithResource(res pcommon.Resource) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	res.CopyTo(rm.Resource())
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName(metadata.ScopeName)
	sm.Scope().SetVersion(sb.settings.BuildInfo.Version)
	return metrics
}

// createMetrics creates metrics for a given metadata element. If value or metric are nil, only resource+status are emitted.
func (sb *brokerSubscription) createMetrics(meta *subscriptionMetadata, value *string) (pmetric.Metrics, error) {
	rb := metadata.NewResourceBuilder(metadata.DefaultResourceAttributesConfig())
	rb.SetSwOtelcolMqttBrokerName(meta.broker.Name)
	rb.SetSwOtelcolMqttBrokerServer(meta.broker.Server)
	rb.SetSwOtelcolMqttBrokerPort(int64(meta.broker.Port))
	switch meta.broker.Protocol {
	case "mqtt":
		rb.SetSwOtelcolMqttBrokerProtocolMqtt()
	case "mqtts":
		rb.SetSwOtelcolMqttBrokerProtocolMqtts()
	}
	if meta.sensor != nil {
		rb.SetSwOtelcolMqttSensorName(meta.sensor.Name)
	}
	res := rb.Emit()
	metrics := sb.newMetricsWithResource(res)
	rm := metrics.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	m := sm.Metrics().AppendEmpty()
	m.SetName(fmt.Sprintf("sw.otelcol.Mqtt.%s", meta.metric.Name))
	m.SetDescription(meta.metric.Desc)
	m.SetUnit(meta.metric.Unit)
	m.SetEmptyGauge()
	dp := m.Gauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	if err := setDataPointValue(dp, *value, meta.metric.Type); err != nil {
		return metrics, err
	}
	return metrics, nil
}

// extractValue extracts the value from the MQTT message payload
func extractValue(message mqtt.Message, jsonProperty string) (string, error) {
	payload := message.Payload()

	if jsonProperty == "" {
		return string(payload), nil
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(payload, &jsonData); err != nil {
		return "", err
	}

	value, ok := jsonData[jsonProperty]
	if !ok {
		return "", fmt.Errorf("JSON property '%s' not found in payload", jsonProperty)
	}

	return fmt.Sprintf("%v", value), nil
}

// setDataPointValue sets the appropriate value type on the data point based on metric type
func setDataPointValue(dataPoint pmetric.NumberDataPoint, strValue string, metricType string) error {
	switch metricType {
	case "int":
		intValue, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse payload to int: %s", strValue)
		}
		dataPoint.SetIntValue(intValue)
	case "float":
		floatValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return fmt.Errorf("failed to parse payload to float: %s", strValue)
		}
		dataPoint.SetDoubleValue(floatValue)
	case "string":
		var intValue int64
		if strValue == "on" || strValue == "true" {
			intValue = 1
		} else {
			intValue = 0
		}
		dataPoint.SetIntValue(intValue)
	default:
		return fmt.Errorf("unsupported metric type: %s", metricType)
	}

	return nil
}

func setBrokerAttributes(rb *metadata.ResourceBuilder, broker *Broker) {
	rb.SetSwOtelcolMqttBrokerName(broker.Name)
	rb.SetSwOtelcolMqttBrokerServer(broker.Server)
	rb.SetSwOtelcolMqttBrokerPort(int64(broker.Port))
	switch broker.Protocol {
	case "mqtt":
		rb.SetSwOtelcolMqttBrokerProtocolMqtt()
	case "mqtts":
		rb.SetSwOtelcolMqttBrokerProtocolMqtts()
	}
}

func setStatus(rb *metadata.ResourceBuilder, status string) {
	switch status {
	case StatusOK:
		rb.SetSwOtelcolMqttBrokerStatusOK()
	case StatusConnectionFailed:
		rb.SetSwOtelcolMqttBrokerStatusConnectionFailed()
	case StatusRoundtripFailed:
		rb.SetSwOtelcolMqttBrokerStatusRoundtripFailed()
	case StatusSubscribeFailed:
		rb.SetSwOtelcolMqttBrokerStatusSubscribeFailed()
	}
}

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
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
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
	broker             *Broker
	client             *mqtt.Client
	logger             *zap.Logger
	consumer           consumer.Metrics
	ctx                context.Context
	wg                 sync.WaitGroup
	topicSubscriptions map[string][]*subscriptionMetadata
}

// newBrokerSubscription creates a new subscribed broker instance
func newBrokerSubscription(broker *Broker, logger *zap.Logger, consumer consumer.Metrics, ctx context.Context) *brokerSubscription {
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

	return &brokerSubscription{
		broker:             broker,
		logger:             logger,
		consumer:           consumer,
		ctx:                ctx,
		topicSubscriptions: topicSubscriptions,
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
	for _, metadata := range metricMetadata {
		err := sb.handleMessage(msg, metadata)
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

	metrics, err := createMetrics(metadata, StatusOK, &value)
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

		sb.logger.Debug("Roundtrip completed",
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

// createRoundtripMetric creates a roundtrip metric without consuming it
func (sb *brokerSubscription) createRoundtripMetric(duration time.Duration) pmetric.Metrics {

	metricMetadata := &Metric{
		Name:         "Roundtrip",
		Type:         "int",
		Topic:        "",
		Unit:         "ms",
		Desc:         "Time taken to publish and receive a message",
		JsonProperty: "",
	}

	value := strconv.FormatInt(duration.Milliseconds(), 10)
	metrics, err := createMetrics(newSubscriptionMetadata(sb.broker, nil, metricMetadata), StatusOK, &value)

	if err != nil {
		sb.logger.Error("Failed to create roundtrip metric",
			zap.String("broker", sb.broker.Name),
			zap.Error(err))
		return pmetric.Metrics{}
	}

	return metrics
}

// sendRoundtripMetric creates and sends a roundtrip metric
func (sb *brokerSubscription) sendRoundtripMetric(duration time.Duration) error {
	return sb.consumer.ConsumeMetrics(sb.ctx, sb.createRoundtripMetric(duration))
}

func (sb *brokerSubscription) sendBrokerStatus(status string) {
	metrics, _ := createMetrics(newSubscriptionMetadata(sb.broker, nil, nil), status, nil)

	if err := sb.consumer.ConsumeMetrics(sb.ctx, metrics); err != nil {
		sb.logger.Error("Failed to send broker status",
			zap.String("broker", sb.broker.Name),
			zap.String("status", status),
			zap.Error(err))
	}
}

// createMetrics creates and populates resource metrics with broker and sensor metadata
func createMetrics(metadata *subscriptionMetadata, status string, value *string) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	rms := metrics.ResourceMetrics()
	rm := rms.AppendEmpty()

	attrs := rm.Resource().Attributes()
	attrs.PutStr("sw.otelcol.IotBroker.name", metadata.broker.Name)
	attrs.PutStr("sw.otelcol.IotBroker.server", metadata.broker.Server)
	attrs.PutInt("sw.otelcol.IotBroker.port", int64(metadata.broker.Port))
	attrs.PutStr("sw.otelcol.IotBroker.protocol", metadata.broker.Protocol)

	if metadata.sensor != nil {
		attrs.PutStr("sw.otelcol.IotSensor.name", metadata.sensor.Name)
		attrs.PutStr("sw.otelcol.IotSensor.category", metadata.sensor.Category)
		attrs.PutStr("sw.otelcol.IotBroker.status", status)
	}

	if value == nil {
		return metrics, nil
	}

	metric := rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	metric.SetName(fmt.Sprintf("sw.otelcol.IoT.%s", metadata.metric.Name))
	metric.SetDescription(metadata.metric.Desc)
	metric.SetUnit(metadata.metric.Unit)
	metric.SetEmptyGauge()

	dataPoint := metric.Gauge().DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	if err := setDataPointValue(dataPoint, *value, metadata.metric.Type); err != nil {
		return metrics, err
	}

	return metrics, nil
}

// addValueToMetrics processes a single metric from the MQTT message
func addValueToMetrics(value string, m *Metric, metrics pmetric.Metrics) error {
	ilm := metrics.ResourceMetrics().At(0).ScopeMetrics().AppendEmpty()
	metric := ilm.Metrics().AppendEmpty()
	metric.SetName(fmt.Sprintf("sw.otelcol.IoT.%s", m.Name))
	metric.SetDescription(m.Desc)
	metric.SetUnit(m.Unit)
	metric.SetEmptyGauge()

	dataPoint := metric.Gauge().DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	return setDataPointValue(dataPoint, value, m.Type)
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

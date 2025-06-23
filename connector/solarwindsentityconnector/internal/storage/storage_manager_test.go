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

package storage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// mockEventConsumer implements the internal.EventConsumer interface for testing
type mockEventConsumer struct {
	mu             sync.Mutex
	receivedEvents [][]internal.Event
	calledTimes    int
}

func (m *mockEventConsumer) SendExpiredEvents(ctx context.Context, events []internal.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Make a copy of the events to avoid potential race conditions
	eventsCopy := make([]internal.Event, len(events))
	copy(eventsCopy, events)
	m.receivedEvents = append(m.receivedEvents, eventsCopy)
	m.calledTimes++
}

// mockEvent implements the internal.Event interface for testing
type mockEvent struct {
	id string
}

func (m *mockEvent) Update(logRecords *plog.LogRecordSlice) {
	// No-op for testing
}

func (m *mockEvent) Delete(logRecords *plog.LogRecordSlice) {
	// No-op for testing
}

func TestReceiveExpired_CanceledContext_ClosesChannel(t *testing.T) {
	// Create a mock consumer
	mockConsumer := &mockEventConsumer{
		receivedEvents: make([][]internal.Event, 0),
	}

	// Create a manager with the mock consumer
	manager := &Manager{
		expiredCh:     make(chan internal.Event),
		eventConsumer: mockConsumer,
		logger:        zap.NewNop(),
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start the receiveExpired function in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.receiveExpired(ctx)
	}()

	// Cancel the context immediately
	cancel()

	// Wait for the goroutine to finish
	wg.Wait()

	// Verify the channel is closed
	_, isOpen := <-manager.expiredCh
	assert.False(t, isOpen, "expiredCh should be closed when context is cancelled")
	// Verify no events were sent
	assert.Equal(t, 0, mockConsumer.calledTimes, "SendExpiredEvents should not be called when context is cancelled")
}

func TestReceiveExpired_MultipleBatches(t *testing.T) {
	// Create a mock consumer
	mockConsumer := &mockEventConsumer{
		receivedEvents: make([][]internal.Event, 0),
	}

	// Create a manager with the mock consumer
	manager := &Manager{
		expiredCh:     make(chan internal.Event),
		eventConsumer: mockConsumer,
		logger:        zap.NewNop(),
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the receiveExpired function in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.receiveExpired(ctx)
	}()

	// Send first batch of events
	for i := 0; i < 3; i++ {
		manager.expiredCh <- &mockEvent{id: "batch1_" + string(rune('0'+i))}
	}

	// Wait for the first batch to be processed
	time.Sleep(1500 * time.Millisecond)

	// Send second batch of events
	for i := 0; i < 2; i++ {
		manager.expiredCh <- &mockEvent{id: "batch2_" + string(rune('0'+i))}
	}

	// Wait for the second batch to be processed
	time.Sleep(1500 * time.Millisecond)

	// Cancel context to stop the goroutine
	cancel()

	// Wait for the goroutine to finish
	wg.Wait()

	// Verify the events were batched and sent correctly
	assert.Equal(t, 2, mockConsumer.calledTimes, "SendExpiredEvents should be called twice")
	assert.Equal(t, 2, len(mockConsumer.receivedEvents), "Two batches should be received")
	assert.Equal(t, 3, len(mockConsumer.receivedEvents[0]), "First batch should contain 3 events")
	assert.Equal(t, 2, len(mockConsumer.receivedEvents[1]), "Second batch should contain 2 events")
}

func TestReceiveExpired_EmptyBatch(t *testing.T) {
	// Create a mock consumer
	mockConsumer := &mockEventConsumer{
		receivedEvents: make([][]internal.Event, 0),
	}

	// Create a manager with the mock consumer
	manager := &Manager{
		expiredCh:     make(chan internal.Event),
		eventConsumer: mockConsumer,
		logger:        zap.NewNop(),
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the receiveExpired function in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.receiveExpired(ctx)
	}()

	// Let the timer expire without sending any events
	time.Sleep(1500 * time.Millisecond)

	// Cancel context to stop the goroutine
	cancel()

	// Wait for the goroutine to finish
	wg.Wait()

	// Verify no events were sent because the batch was empty
	assert.Equal(t, 0, mockConsumer.calledTimes, "SendExpiredEvents should not be called for empty batch")
	assert.Equal(t, 0, len(mockConsumer.receivedEvents), "No batches should be received")
}

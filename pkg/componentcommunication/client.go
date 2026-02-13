package componentcommunication

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampcustommessages"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type Message struct {
	CustomMessage *protobufs.CustomMessage
	ExtensionID   string
}

type Client interface {
	Start(ctx context.Context, host component.Host) (<-chan *Message, <-chan error)
	Stop()
}

type client struct {
	logger *zap.Logger
	config Config

	errors   chan error
	messages chan *Message

	handlersArea sync.Mutex
	handlers     map[string]opampcustommessages.CustomCapabilityHandler

	contextArea     sync.Mutex
	messagingCtx    context.Context
	cancelMessaging context.CancelFunc

	running   atomic.Bool // Single run.
	runningWg *sync.WaitGroup
}

func NewClient(logger *zap.Logger, config Config) (Client, error) {
	if err := validateConfig(logger, config); err != nil {
		msg := "invalid configuration for component communication client"
		logger.Error(msg, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return &client{
		config:   config,
		handlers: make(map[string]opampcustommessages.CustomCapabilityHandler, len(config.Capabilities)),
		logger:   logger,
	}, nil
}

func validateConfig(logger *zap.Logger, config Config) error {
	if config.OpAMPExtensionID == "" {
		msg := "opamp_extension must be defined"
		logger.Error(msg)
		return errors.New(msg)
	}

	if len(config.Capabilities) <= 0 {
		msg := "capabilities must be defined"
		logger.Error(msg)
		return errors.New(msg)
	}

	return nil
}

func (c *client) Start(ctx context.Context, host component.Host) (<-chan *Message, <-chan error) {
	if c.running.CompareAndSwap(false, true) {
		c.contextArea.Lock()

		c.errors = make(chan error, 1)
		c.messages = make(chan *Message, len(c.config.Capabilities))

		// Create context for this run.
		c.messagingCtx, c.cancelMessaging = context.WithCancel(ctx)

		err := c.startInternal(host)
		if err != nil {
			msg := "failed to start component communication client"
			c.logger.Error(msg, zap.Error(err))
			c.errors <- fmt.Errorf("%s: %w", msg, err)

			// Nothing will come from communication client, so we can close channels immediately.
			close(c.errors)
			close(c.messages)
			return c.messages, c.errors
		}

		c.contextArea.Unlock()
	} else {
		msg := "controller is already running"
		c.logger.Error(msg)
		c.errors <- errors.New(msg)
	}

	return c.messages, c.errors
}

func (c *client) startInternal(host component.Host) error {
	if err := c.registerCommunication(host); err != nil {
		msg := "failed to register component communication"
		c.logger.Error(msg, zap.Error(err))
		return fmt.Errorf("%s: %w", msg, err)
	}

	c.runningWg = new(sync.WaitGroup)
	c.runningWg.Add(1)

	receivingLoopsStarted := new(sync.WaitGroup)
	receivingLoopsStarted.Add(1)

	go c.spinReceivingLoops(receivingLoopsStarted)

	// All routines are ready. Messaging is ready.
	receivingLoopsStarted.Wait()

	return nil
}

func (c *client) registerCommunication(host component.Host) error {
	registry, err := extensionfinder.FindExtension[opampcustommessages.CustomCapabilityRegistry](c.logger, c.config.OpAMPExtensionID, host)
	if err != nil {
		msg := "failed to find OpAMP extension"
		c.logger.Error(msg, zap.Error(err))
		return fmt.Errorf("%s: %w", msg, err)
	}

	for _, capability := range c.config.Capabilities {
		c.logger.Info(
			"Registering capability into OpAMP extension",
			zap.String("capability", capability),
			zap.String("extension", c.config.OpAMPExtensionID))

		handler, err := registry.Register(capability)
		if err != nil {
			msg := "failed to register capability"
			c.logger.Error(msg, zap.Error(err))
			return fmt.Errorf("%s: %w", msg, err)
		}
		c.handlers[capability] = handler
	}

	return nil
}

func (c *client) spinReceivingLoops(receivingLoopsReadiness *sync.WaitGroup) {
	// All receiving routines are done.
	defer c.runningWg.Done()
	// Receiving loops are done, close aggregating channel.
	defer close(c.messages)

	// Create wait group for this run.
	receivingLoopsWg := new(sync.WaitGroup)

	for capability, handler := range c.handlers {
		receivingLoopsWg.Add(1)
		receivingLoopsReadiness.Add(1)

		go c.receivingLoop(c.messagingCtx, capability, handler, receivingLoopsWg, receivingLoopsReadiness)
	}

	// Started.
	receivingLoopsReadiness.Done()

	// Wait until all receiving loops are done, then unregister communication.
	receivingLoopsWg.Wait()
}

func (c *client) receivingLoop(
	ctx context.Context,
	capability string,
	handler opampcustommessages.CustomCapabilityHandler,
	receivingLoopsWg *sync.WaitGroup,
	receivingLoopsReadiness *sync.WaitGroup,
) {
	// Receiving loop is done.
	defer receivingLoopsWg.Done()

	// Unregister capability from extension when the loop is done.
	defer func() {
		c.logger.Info(
			"Unregistering capability from OpAMP extension",
			zap.String("capability", capability),
			zap.String("extension", c.config.OpAMPExtensionID))

		handler.Unregister()
	}()

	c.logger.Info(
		"Starting receiving loop for capability from OpAMP extension",
		zap.String("capability", capability),
		zap.String("extension", c.config.OpAMPExtensionID))

	// Started.
	receivingLoopsReadiness.Done()

	for {
		select {
		// Context watching.
		case <-ctx.Done():
			c.logger.Info(
				"Stopping receiving loop for capability by context cancellation",
				zap.String("capability", capability),
				zap.String("extension", c.config.OpAMPExtensionID))
			return

		// Messages watching.
		case msg, ok := <-handler.Message():
			if !ok {
				c.logger.Info(
					"Handler message channel closed for capability from OpAMP extension",
					zap.String("capability", capability),
					zap.String("extension", c.config.OpAMPExtensionID))
				return
			}

			c.logger.Info(
				"Received message for capability from OpAMP extension",
				zap.String("capability", capability),
				zap.String("extension", c.config.OpAMPExtensionID))

			c.messages <- &Message{
				CustomMessage: msg,
				ExtensionID:   c.config.OpAMPExtensionID,
			}
		}
	}
}

func (c *client) Stop() {
	if !c.running.CompareAndSwap(true, false) {
		c.logger.Info("Stopping component communication client for extension", zap.String("extension", c.config.OpAMPExtensionID))

		c.contextArea.Lock()

		if c.cancelMessaging != nil {
			c.cancelMessaging()
			c.cancelMessaging = nil
		}

		if c.runningWg != nil {
			c.runningWg.Wait()
			c.runningWg = nil
		}

		c.contextArea.Unlock()
	} else {
		c.logger.Debug("component communication client is not running, nothing to stop")
	}
}

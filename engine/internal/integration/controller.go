package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	Config     configkit.RichIntegration
	MessageIDs *envelope.MessageIDGenerator
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() configkit.RichHandler {
	return c.Config
}

// Tick does nothing.
func (c *Controller) Tick(
	context.Context,
	fact.Observer,
	time.Time,
) ([]*envelope.Envelope, error) {
	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	mt := message.TypeOf(env.Message)

	if !c.Config.MessageTypes()[mt].IsConsumed {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), mt))
	}

	s := &scope{
		config:     c.Config,
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		command:    env,
	}

	var err error
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"IntegrationMessageHandler",
		"HandleCommand",
		c.Config.Handler(),
		env.Message,
		func() {
			err = c.Config.Handler().HandleCommand(
				ctx,
				s,
				env.Message.(dogma.Command),
			)
		},
	)

	return s.events, err
}

// Reset does nothing.
func (c *Controller) Reset() {
}

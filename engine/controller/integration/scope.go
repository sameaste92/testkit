package integration

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.IntegrationCommandScope.
type scope struct {
	config     configkit.RichIntegration
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	now        time.Time
	produced   message.TypeCollection
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.produced.HasM(m) {
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to record events of type %T",
			s.config.Identity().Name,
			m,
		))
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not record event of type %T, it is invalid: %s",
			m,
			err,
		))
	}

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			HandlerName: s.config.Identity().Name,
			HandlerType: configkit.IntegrationHandlerType,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		HandlerName:   s.config.Identity().Name,
		Handler:       s.config.Handler(),
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		HandlerName:  s.config.Identity().Name,
		Handler:      s.config.Handler(),
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}

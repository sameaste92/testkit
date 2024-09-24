package aggregate_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *AggregateMessageHandlerStub
		config     configkit.RichAggregate
		ctrl       *Controller
		command    *envelope.Envelope
	)

	g.BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			CommandA1,
			time.Now(),
		)

		handler = &AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<name>", "fd88e430-32fe-49a6-888f-f678dcf924ef")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
			RouteCommandToInstanceFunc: func(m dogma.Command) string {
				switch m.(type) {
				case CommandStub[TypeA]:
					return "<instance>"
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		config = configkit.FromAggregate(handler)

		ctrl = &Controller{
			Config:     config,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	g.When("the instance does not exist", func() {
		g.Describe("func Destroy()", func() {
			g.It("does not record a fact", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.Destroy()
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.AggregateInstanceDestroyed{}),
				))
			})
		})

		g.Describe("func RecordEvent()", func() {
			g.It("records facts about instance creation and the event", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventA1)
				}

				now := time.Now()
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceCreated{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
					},
				))
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     config,
								HandlerType: configkit.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})
		})
	})

	g.When("the instance exists", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1) // record event to create the instance
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				envelope.NewCommand(
					"2000",
					CommandA2, // use a different message to create the instance
					time.Now(),
				),
			)

			Expect(err).ShouldNot(HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		g.Describe("func Destroy()", func() {
			g.It("records a fact", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.Destroy()
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceDestroyed{
						Handler:    config,
						InstanceID: "<instance>",
						Root:       &AggregateRootStub{},
						Envelope:   command,
					},
				))
			})
		})

		g.Describe("func RecordEvent()", func() {
			g.BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventA1)
				}
			})

			g.It("records a fact", func() {
				messageIDs.Reset() // reset after setup for a predictable ID.

				buf := &fact.Buffer{}
				now := time.Now()
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
								EventA1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     config,
								HandlerType: configkit.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})

			g.It("does not record a fact about instance creation", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.AggregateInstanceDestroyed{}),
				))
			})

			g.It("records facts about reverting destruction and the event if called after Destroy()", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.Destroy()
					s.RecordEvent(EventA1)
				}

				now := time.Now()
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceDestructionReverted{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
					},
				))
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     config,
								HandlerType: configkit.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})

			g.It("panics if the event type is not configured to be produced", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventX1)
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        Equal(config),
							"Interface":      Equal("AggregateMessageHandler"),
							"Method":         Equal("HandleCommand"),
							"Implementation": Equal(config.Handler()),
							"Message":        Equal(command.Message),
							"Description":    Equal("recorded an event of type stubs.EventStub[TypeX], which is not produced by this handler"),
							"Location": MatchAllFields(
								Fields{
									"Func": Not(BeEmpty()),
									"File": HaveSuffix("/engine/internal/aggregate/scope_test.go"),
									"Line": Not(BeZero()),
								},
							),
						},
					),
				))
			})

			g.It("panics if the event is invalid", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventStub[TypeA]{
						ValidationError: "<invalid>",
					})
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        Equal(config),
							"Interface":      Equal("AggregateMessageHandler"),
							"Method":         Equal("HandleCommand"),
							"Implementation": Equal(config.Handler()),
							"Message":        Equal(command.Message),
							"Description":    Equal("recorded an invalid stubs.EventStub[TypeA] event: <invalid>"),
							"Location": MatchAllFields(
								Fields{
									"Func": Not(BeEmpty()),
									"File": HaveSuffix("/engine/internal/aggregate/scope_test.go"),
									"Line": Not(BeZero()),
								},
							),
						},
					),
				))
			})
		})
	})

	g.Describe("func InstanceID()", func() {
		g.It("returns the instance ID", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				called = true
				Expect(s.InstanceID()).To(Equal("<instance>"))
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})

	g.Describe("func Log()", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
			}
		})

		g.It("records a fact", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByAggregate{
					Handler:    config,
					InstanceID: "<instance>",
					Root:       &AggregateRootStub{},
					Envelope:   command,
					LogFormat:  "<format>",
					LogArguments: []any{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})

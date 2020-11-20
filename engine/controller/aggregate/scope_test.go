package aggregate_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller/aggregate"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *AggregateMessageHandler
		config     configkit.RichAggregate
		ctrl       *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			MessageA1,
			time.Now(),
		)

		handler = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
			RouteCommandToInstanceFunc: func(m dogma.Message) string {
				switch m.(type) {
				case MessageA:
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

	When("the instance does not exist", func() {
		Describe("func Destroy()", func() {
			It("does not record a fact", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
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

		Describe("func RecordEvent()", func() {
			It("records facts about instance creation and the event", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageE1)
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
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
						Envelope: command,
					},
				))
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							MessageE1,
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

	When("the instance exists", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.RecordEvent(MessageE1) // record event to create the instance
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				envelope.NewCommand(
					"2000",
					MessageA2, // use a different message to create the instance
					time.Now(),
				),
			)

			Expect(err).ShouldNot(HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		Describe("func Destroy()", func() {
			It("records a fact", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
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
						Root:       &AggregateRoot{},
						Envelope:   command,
					},
				))
			})
		})

		Describe("func RecordEvent()", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageE1)
				}
			})

			It("records a fact", func() {
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
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
								MessageE1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							MessageE1,
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

			It("does not record a fact about instance creation", func() {
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

			It("records facts about instance creation and the event if called after Destroy()", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Destroy()
					s.RecordEvent(MessageE1)
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
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
						Envelope: command,
					},
				))
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    config,
						InstanceID: "<instance>",
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							MessageE1,
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

			It("panics if the event type is not configured to be produced", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageX1)
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(PanicWith("the '<name>' handler is not configured to record events of type fixtures.MessageX"))
			})

			It("panics if the event is invalid", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageE{
						Value: errors.New("<invalid>"),
					})
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(PanicWith("can not record event of type fixtures.MessageE, it is invalid: <invalid>"))
			})
		})
	})

	Describe("func InstanceID()", func() {
		It("returns the instance ID", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Message,
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

	Describe("func Log()", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
			}
		})

		It("records a fact", func() {
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
					Root:       &AggregateRoot{},
					Envelope:   command,
					LogFormat:  "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})

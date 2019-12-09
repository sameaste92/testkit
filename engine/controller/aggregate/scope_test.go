package aggregate_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
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
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			MessageA1,
			time.Now(),
		)

		handler = &AggregateMessageHandler{
			RouteCommandToInstanceFunc: func(m dogma.Message) string {
				switch m.(type) {
				case MessageA:
					return "<instance>"
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		controller = NewController(
			configkit.MustNewIdentity("<name>", "<key>"),
			handler,
			&messageIDs,
			message.NewTypeSet(
				MessageBType,
				MessageEType,
			),
		)

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	When("the instance does not exist", func() {
		Describe("func Root()", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Root()
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})

		Describe("func Create()", func() {
			It("returns true", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Create()).To(BeTrue())
					s.RecordEvent(MessageE1) // event must be recorded when creating
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
					s.RecordEvent(MessageE1) // event must be recorded when creating
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceCreated{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &AggregateRoot{},
						Envelope:    command,
					},
				))
			})
		})

		Describe("func Destroy()", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Destroy()
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})

		Describe("func RecordEvent()", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageB1)
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})
	})

	When("the instance exists", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Create()
				s.RecordEvent(MessageE1) // event must be recorded when creating
			}

			_, err := controller.Handle(
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

		Describe("func Root()", func() {
			It("returns the root", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Root()).To(Equal(&AggregateRoot{}))
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Describe("func Create()", func() {
			It("returns false", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Create()).To(BeFalse())
					s.RecordEvent(MessageE1) // event must be recorded when creating
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("does not record a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
					s.RecordEvent(MessageE1) // event must be recorded when creating
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.AggregateInstanceCreated{}),
				))
			})
		})

		Describe("func Destroy()", func() {
			It("records a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(MessageE1) // event must be recorded when destroying
					s.Destroy()
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceDestroyed{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &AggregateRoot{},
						Envelope:    command,
					},
				))
			})
		})

		Describe("func RecordEvent()", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					m dogma.Message,
				) {
					s.RecordEvent(MessageB1)
				}
			})

			It("records a fact", func() {
				messageIDs.Reset() // reset after setup for a predictable ID.

				buf := &fact.Buffer{}
				now := time.Now()
				_, err := controller.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &AggregateRoot{},
						Envelope:    command,
						EventEnvelope: command.NewEvent(
							"1",
							MessageB1,
							now,
							envelope.Origin{
								HandlerName: "<name>",
								HandlerType: configkit.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})

			It("panics if the event type is not configured to be produced", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					m dogma.Message,
				) {
					s.RecordEvent(MessageZ1)
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})
	})

	Describe("func InstanceID()", func() {
		It("returns the instance ID", func() {
			called := false
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				called = true
				Expect(s.InstanceID()).To(Equal("<instance>"))
			}

			_, err := controller.Handle(
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
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
			}
		})

		It("records a fact", func() {
			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByAggregate{
					HandlerName: "<name>",
					Handler:     handler,
					InstanceID:  "<instance>",
					Root:        &AggregateRoot{},
					Envelope:    command,
					LogFormat:   "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})

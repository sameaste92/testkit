package testkit_test

import (
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type BDD Test", func() {
	g.Describe("func Test()", func() {
		g.It("fails the test if the scenario setup returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "066f3ca2-8b0d-4095-8f0a-121bf138971c")
				},
			}

			t := &testingmock.T{FailSilently: true}

			SetupScenario(
				"<scenario>",
				noopAction{errors.New("<error>")},
			).Given(
				"<given>",
				noopAction{nil},
			).When(
				"<when>",
				noopAction{nil},
			).Then(
				"<then>",
				pass,
			).Test(t, app)

			gm.Expect(t.Failed()).To(gm.BeTrue())
		})

		g.It("fails the test if given returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "066f3ca2-8b0d-4095-8f0a-121bf138971c")
				},
			}

			t := &testingmock.T{FailSilently: true}

			SetupScenario(
				"<scenario>",
				noopAction{nil},
			).Given(
				"<given>",
				noopAction{errors.New("<error>")},
			).When(
				"<when>",
				noopAction{nil},
			).Then(
				"<then>",
				pass,
			).Test(t, app)

			gm.Expect(t.Failed()).To(gm.BeTrue())
		})

		g.It("fails the test if when returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "066f3ca2-8b0d-4095-8f0a-121bf138971c")
				},
			}

			t := &testingmock.T{FailSilently: true}

			SetupScenario(
				"<scenario>",
				noopAction{nil},
			).Given(
				"<given>",
				noopAction{nil},
			).When(
				"<when>",
				noopAction{errors.New("<error>")},
			).Then(
				"<then>",
				pass,
			).Test(t, app)

			gm.Expect(t.Failed()).To(gm.BeTrue())
		})

		g.It("fails the test if then returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "066f3ca2-8b0d-4095-8f0a-121bf138971c")
				},
			}

			t := &testingmock.T{FailSilently: true}

			SetupScenario(
				"<scenario>",
				noopAction{nil},
			).Given(
				"<given>",
				noopAction{nil},
			).When(
				"<when>",
				noopAction{nil},
			).Then(
				"<then>",
				fail,
			).Test(t, app)

			gm.Expect(t.Failed()).To(gm.BeTrue())
		})
	})
})

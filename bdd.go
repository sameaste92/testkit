package testkit

import (
	"fmt"
	"strings"

	"github.com/dogmatiq/dapper"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/engine"
)

type scenario struct {
	desc  string
	setup []Action
	given Given
	when  When
	then  Then
}

type Given struct {
	Desc   string
	Action Action
}

type When struct {
	Desc   string
	Action Action
}

type Then struct {
	Desc   string
	Result Expectation
}

// Setup -
func SetupScenario(desc string, actions ...Action) *scenario {
	return &scenario{
		desc:  desc,
		setup: actions,
	}
}

// Given sets an action that doesn't make any expectations. It is
// used to place the application into a particular state.
func (s *scenario) Given(desc string, action Action) *scenario {
	s.given.Desc = desc
	s.given.Action = action

	return s
}

// When sets an action as part of an expectation.
func (s *scenario) When(desc string, action Action) *scenario {
	s.when.Desc = desc
	s.when.Action = action

	return s
}

// Then sets the result as part of an expectation.
func (s *scenario) Then(desc string, expectation Expectation) *scenario {
	s.then.Desc = desc
	s.then.Result = expectation

	return s
}

// Test runs a scenario test
func (s *scenario) Test(
	t TestingT,
	app dogma.Application,
	options ...TestOption,
) *Test {
	testkit := Begin(t, app, options...)

	s.doSetup(testkit)
	s.doGiven(testkit)
	s.doWhenThen(testkit)

	return testkit
}

func (s *scenario) doSetup(testkit *Test) *Test {
	logf(testkit.testingT, "--- SCENARIO %s ---", s.desc)
	for _, act := range s.setup {
		if err := testkit.doAction(act); err != nil {
			testkit.testingT.Fatal(err)
		}
	}

	return testkit
}

func (s *scenario) doGiven(testkit *Test) *Test {
	logf(testkit.testingT, "--- GIVEN %s ---", s.given.Desc)
	if err := testkit.doAction(s.given.Action); err != nil {
		testkit.testingT.Fatal(err)
	}

	return testkit
}

func (s *scenario) doWhenThen(testkit *Test) *Test {
	scope := PredicateScope{
		App:     testkit.app,
		Options: testkit.predicateOptions,
	}

	s.when.Action.ConfigurePredicate(&scope.Options)

	logf(testkit.testingT, "--- WHEN %s THEN %s ---", s.when.Desc, s.then.Desc)

	p, err := s.then.Result.Predicate(scope)
	if err != nil {
		testkit.testingT.Fatal(err)
		return testkit // required when using a mock testingT that does not panic
	}

	// Using a defer inside a closure satisfies the requirements of the
	// Expectation and Predicate interfaces which state that p.Done() must
	// be called exactly once, and that it must be called before calling
	// p.Report().
	if err := func() error {
		defer p.Done()
		return testkit.doAction(s.when.Action, engine.WithObserver(p))
	}(); err != nil {
		testkit.testingT.Fatal(err)
		return testkit // required when using a mock testingT that does not panic
	}

	options := []dapper.Option{
		dapper.WithPackagePaths(false),
		dapper.WithUnexportedStructFields(false),
	}

	ctx := ReportGenerationContext{
		TreeOk:  p.Ok(),
		printer: dapper.NewPrinter(options...),
	}

	rep := p.Report(ctx)

	buf := &strings.Builder{}
	fmt.Fprint(buf, "--- TEST REPORT ---\n\n")
	must.WriteTo(buf, rep)
	testkit.testingT.Log(buf.String())

	if !ctx.TreeOk {
		testkit.testingT.FailNow()
	}

	return testkit
}

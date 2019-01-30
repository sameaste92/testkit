package assert

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// EventRecorded asserts that a specific event is recorded.
type EventRecorded struct {
	messageAssertionBehavior

	Expected dogma.Message
}

// Begin is called before the test is executed.
//
// c is the comparator used to compare messages and other entities.
func (a *EventRecorded) Begin(c compare.Comparator) {
	// reset everything
	a.messageAssertionBehavior = messageAssertionBehavior{
		expected: a.Expected,
		role:     message.EventRole,
		cmp:      c,
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *EventRecorded) End(r render.Renderer) *Result {
	res := &Result{
		Ok: a.ok,
		Criteria: fmt.Sprintf(
			"record a specific '%s' event",
			message.TypeOf(a.Expected),
		),
	}

	if !a.ok {
		if a.best == nil {
			a.buildResultNoMatch(r, res)
		} else {
			a.buildResult(r, res)
		}
	}

	return res
}

// buildResultNoMatch builds the assertion result when there is no "best-match"
// message.
func (a *EventRecorded) buildResultNoMatch(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	if !a.enabledHandlers[handler.AggregateType] &&
		!a.enabledHandlers[handler.IntegrationType] {
		res.Explanation = "no relevant handler types (aggregate and integration) were enabled"
		s.AppendListItem("enable the relevant handler types using the EnableHandlerType() option")
		return
	}

	if !a.enabledHandlers[handler.AggregateType] {
		s.AppendListItem("enable aggregate handlers using the EnableHandlerType() option")
	}

	if !a.enabledHandlers[handler.IntegrationType] {
		s.AppendListItem("enable integration handlers using the EnableHandlerType() option")
	}

	if len(a.engagedHandlers) == 0 {
		res.Explanation = "no relevant handlers (aggregates or integrations) were engaged"
		s.AppendListItem("check the application's routing configuration")
		return
	}

	if a.commands == 0 && a.events == 0 {
		res.Explanation = "no messages were produced at all"
	} else if a.events == 0 {
		res.Explanation = "no events were recorded at all"
	} else {
		res.Explanation = "none of the engaged handlers recorded the expected event"
	}

	for n, t := range a.engagedHandlers {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t)
	}
}

// buildResultNoMatch builds the assertion result when there is a "best-match"
// message available.
func (a *EventRecorded) buildResult(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	// the "best match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		res.Explanation = fmt.Sprintf(
			"the expected message was executed as a command by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem(
			"verify that the '%s' %s message handler intended to execute a command of this type",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem("verify that EventRecorded is the correct assertion, did you mean CommandExecuted?")
		return
	}

	if a.sim == compare.SameTypes {
		res.Explanation = fmt.Sprintf(
			"a similar event was recorded by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		s.AppendListItem("check the content of the message")
	} else {
		res.Explanation = fmt.Sprintf(
			"an event of a similar type was recorded by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	render.WriteDiff(
		&res.Section(diffSection).Content,
		render.Message(r, a.Expected),
		render.Message(r, a.best.Message),
	)
}

package dogmatest

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/assert"
)

// CommandExecuted returns an assertion that passes if m is executed as a command.
func CommandExecuted(m dogma.Message) assert.Assertion {
	return &assert.CommandExecuted{
		Expected: m,
	}
}

// EventRecorded returns an assertion that passes if m is recorded as an event.
func EventRecorded(m dogma.Message) assert.Assertion {
	return &assert.EventRecorded{
		Expected: m,
	}
}

// // CommandTypeExecuted returns an assertion that passes if a message with the
// // same type as m is executed as a command.
// func CommandTypeExecuted(m dogma.Message) assert.Assertion {
// 	return &assert.MessageTypeAssertion{
// 		Type: message.TypeOf(m),
// 		Role: message.CommandRole,
// 	}
// }

// // EventTypeRecorded returns an assertion that passes if a message witn the same
// // type as m is recorded as an event.
// func EventTypeRecorded(m dogma.Message) assert.Assertion {
// 	return &assert.MessageTypeAssertion{
// 		Type: message.TypeOf(m),
// 		Role: message.EventRole,
// 	}
// }

// AllOf returns an assertion that passes if all of the given sub-assertions pass.
func AllOf(subs ...assert.Assertion) assert.Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	if n == 1 {
		return subs[0]
	}

	return &assert.CompositeAssertion{
		Criteria:      "all of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			n := len(subs)

			if p == n {
				return "", true
			}

			return fmt.Sprintf(
				"%d of the sub-assertions failed",
				n-p,
			), false
		},
	}
}

// AnyOf returns an assertion that passes if at least one of the given
// sub-assertions passes.
func AnyOf(subs ...assert.Assertion) assert.Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	if n == 1 {
		return subs[0]
	}

	return &assert.CompositeAssertion{
		Criteria:      "any of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			if p > 0 {
				return "", true
			}

			return fmt.Sprintf(
				"all %d of the sub-assertions failed",
				n,
			), false
		},
	}
}

// NoneOf returns an assertion that passes if all of the given sub-assertions
// fail.
func NoneOf(subs ...assert.Assertion) assert.Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	return &assert.CompositeAssertion{
		Criteria:      "none of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			if p == 0 {
				return "", true
			}

			if n == 1 {
				return "the sub-assertion passed unexpectedly", false
			}

			return fmt.Sprintf(
				"%d of the sub-assertions passed unexpectedly",
				p,
			), false
		},
	}
}

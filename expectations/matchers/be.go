package matchers

import (
	"fmt"

	"github.com/drborges/rivers/expectations"
)

// Be provides an expectations.MatchFunc that verifies whether the value under
// test holds the same reference as the expected one.
func Be(expected interface{}) expectations.MatchFunc {
	return func(actual interface{}) error {
		if actual != expected {
			return fmt.Errorf("Expected %v, to be %v", expected, actual)
		}

		return nil
	}
}

package runtime

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/spcent/x/testutil"
)

func TestMust(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail struct {
		err error
	}
	type thenExpected struct {
		shouldPanic bool
	}

	tests := []testutil.Case[givenDetail, whenDetail, thenExpected]{
		{
			Scenario: "Must with nil error",
			When:     "calling Must with nil error",
			Then:     "should not panic",
			WhenDetail: whenDetail{
				err: nil,
			},
			ThenExpected: thenExpected{
				shouldPanic: false,
			},
		},
		{
			Scenario: "Must with non-nil error",
			When:     "calling Must with non-nil error",
			Then:     "should panic",
			WhenDetail: whenDetail{
				err: errors.New("test error"),
			},
			ThenExpected: thenExpected{
				shouldPanic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// When & Then.
			if tt.ThenExpected.shouldPanic {
				Expect(func() { Must(tt.WhenDetail.err) }).To(Panic())
			} else {
				Expect(func() { Must(tt.WhenDetail.err) }).NotTo(Panic())
			}
		})
	}
}

func TestMust1(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail struct {
		value string
		err   error
	}
	type thenExpected struct {
		value       string
		shouldPanic bool
	}

	tests := []testutil.Case[givenDetail, whenDetail, thenExpected]{
		{
			Scenario: "Must1 with nil error",
			When:     "calling Must1 with a value and nil error",
			Then:     "should return the value without panic",
			WhenDetail: whenDetail{
				value: "test value",
				err:   nil,
			},
			ThenExpected: thenExpected{
				value:       "test value",
				shouldPanic: false,
			},
		},
		{
			Scenario: "Must1 with non-nil error",
			When:     "calling Must1 with a value and non-nil error",
			Then:     "should panic",
			WhenDetail: whenDetail{
				value: "test value",
				err:   errors.New("test error"),
			},
			ThenExpected: thenExpected{
				shouldPanic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// When & Then.
			if tt.ThenExpected.shouldPanic {
				Expect(func() { Must1(tt.WhenDetail.value, tt.WhenDetail.err) }).To(Panic())
			} else {
				result := Must1(tt.WhenDetail.value, tt.WhenDetail.err)
				Expect(result).To(Equal(tt.ThenExpected.value))
			}
		})
	}
}

func TestMust2(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail struct {
		value1 string
		value2 int
		err    error
	}
	type thenExpected struct {
		value1      string
		value2      int
		shouldPanic bool
	}

	tests := []testutil.Case[givenDetail, whenDetail, thenExpected]{
		{
			Scenario: "Must2 with nil error",
			When:     "calling Must2 with two values and nil error",
			Then:     "should return both values without panic",
			WhenDetail: whenDetail{
				value1: "test value",
				value2: 42,
				err:    nil,
			},
			ThenExpected: thenExpected{
				value1:      "test value",
				value2:      42,
				shouldPanic: false,
			},
		},
		{
			Scenario: "Must2 with non-nil error",
			When:     "calling Must2 with two values and non-nil error",
			Then:     "should panic",
			WhenDetail: whenDetail{
				value1: "test value",
				value2: 42,
				err:    errors.New("test error"),
			},
			ThenExpected: thenExpected{
				shouldPanic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// When & Then.
			if tt.ThenExpected.shouldPanic {
				Expect(func() {
					Must2(tt.WhenDetail.value1, tt.WhenDetail.value2, tt.WhenDetail.err)
				}).To(Panic())
			} else {
				result1, result2 := Must2(tt.WhenDetail.value1, tt.WhenDetail.value2, tt.WhenDetail.err)
				Expect(result1).To(Equal(tt.ThenExpected.value1))
				Expect(result2).To(Equal(tt.ThenExpected.value2))
			}
		})
	}
}

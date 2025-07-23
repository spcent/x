package time

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/spcent/x/testutil"
)

func TestInRange(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail struct {
		t     time.Time
		start time.Time
		end   time.Time
	}
	type thenExpected struct {
		result bool
	}

	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	before := baseTime.Add(-1 * time.Hour)
	after := baseTime.Add(1 * time.Hour)
	muchAfter := baseTime.Add(2 * time.Hour)

	tests := []testutil.Case[givenDetail, whenDetail, thenExpected]{
		{
			Scenario: "Time is in range",
			When:     "a time that is between start and end, checking if time is in range",
			Then:     "should return true",
			WhenDetail: whenDetail{
				t:     baseTime,
				start: before,
				end:   after,
			},
			ThenExpected: thenExpected{
				result: true,
			},
		},
		{
			Scenario: "Time equals start",
			When:     "a time that equals the start time, checking if time is in range",
			Then:     "should return false",
			WhenDetail: whenDetail{
				t:     before,
				start: before,
				end:   after,
			},
			ThenExpected: thenExpected{
				result: false,
			},
		},
		{
			Scenario: "Time equals end",
			When:     "a time that equals the end time, checking if time is in range",
			Then:     "should return false",
			WhenDetail: whenDetail{
				t:     after,
				start: before,
				end:   after,
			},
			ThenExpected: thenExpected{
				result: false,
			},
		},
		{
			Scenario: "Time is before range",
			When:     "a time that is before the start time, checking if time is in range",
			Then:     "should return false",
			WhenDetail: whenDetail{
				t:     before,
				start: baseTime,
				end:   after,
			},
			ThenExpected: thenExpected{
				result: false,
			},
		},
		{
			Scenario: "Time is after range",
			When:     "a time that is after the end time, checking if time is in range",
			Then:     "should return false",
			WhenDetail: whenDetail{
				t:     muchAfter,
				start: before,
				end:   after,
			},
			ThenExpected: thenExpected{
				result: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// When.
			result := InRange(tt.WhenDetail.t, tt.WhenDetail.start, tt.WhenDetail.end)

			// Then.
			Expect(result).To(Equal(tt.ThenExpected.result))
		})
	}
}

func TestFormat(t *testing.T) {
	pt, err := time.Parse(FormatLayout, "2019-08-03 19:09:08")
	assert.NoError(t, err)
	assert.Equal(t, "2019-08-03 19:09:08", Format(pt, "YYYY-MM-DD HH:mm:ss"))
	assert.Equal(t, "19-08-03 19:09:08", Format(pt, "YY-MM-DD HH:mm:ss"))
	assert.Equal(t, "19-08-03 07:09:08", Format(pt, "YY-MM-DD hh:mm:ss"))
	assert.Equal(t, "August 03，2019", Format(pt, "MMMM DD，YYYY"))
	assert.Equal(t, "Aug 03，2019", Format(pt, "MMM DD，YYYY"))
	assert.Equal(t, "7:9:8", Format(pt, "h:m:s"))
}

func TestParse(t *testing.T) {
	pt, err := Parse(FormatLayout, "2019-08-03 19:09:08")
	assert.NoError(t, err)
	assert.Equal(t, "2019-08-03 19:09:08", Format(pt, "YYYY-MM-DD HH:mm:ss"))
}

func TestShrink(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("1s"))
	if err != nil {
		t.Fatalf("TestShrink:  d.UnmarshalText failed!err:=%v", err)
	}
	c := context.Background()
	to, ctx, cancel := d.Shrink(c)
	defer cancel()
	if time.Duration(to) != time.Second {
		t.Fatalf("new timeout must be equal 1 second")
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second || time.Until(deadline) < time.Millisecond*500 {
		t.Fatalf("ctx deadline must be less than 1s and greater than 500ms")
	}
}

func TestShrinkWithTimeout(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("1s"))
	if err != nil {
		t.Fatalf("TestShrink:  d.UnmarshalText failed!err:=%v", err)
	}
	c, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	to, ctx, cancel := d.Shrink(c)
	defer cancel()
	if time.Duration(to) != time.Second {
		t.Fatalf("new timeout must be equal 1 second")
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second || time.Until(deadline) < time.Millisecond*500 {
		t.Fatalf("ctx deadline must be less than 1s and greater than 500ms")
	}
}

func TestShrinkWithDeadline(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("1s"))
	if err != nil {
		t.Fatalf("TestShrink:  d.UnmarshalText failed!err:=%v", err)
	}
	c, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	to, ctx, cancel := d.Shrink(c)
	defer cancel()
	if time.Duration(to) >= time.Millisecond*500 {
		t.Fatalf("new timeout must be less than 500 ms")
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Millisecond*500 || time.Until(deadline) < time.Millisecond*200 {
		t.Fatalf("ctx deadline must be less than 500ms and greater than 200ms")
	}
}

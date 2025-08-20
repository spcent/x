package time

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	xtime "time"
)

const DateFormat = "2006-01-02"
const FormatLayout = "2006-01-02 15:04:05"

const (
	Day   = 24 * xtime.Hour
	Week  = 7 * Day
	Month = 30 * Day
	Year  = 365 * Day
)

type Item struct {
	Go  string
	Std string
}

var formats = []Item{
	{Std: "YYYY", Go: "2006"},
	{Std: "YY", Go: "06"},
	{Std: "MMMM", Go: "January"},
	{Std: "MMM", Go: "Jan"},
	{Std: "MM", Go: "01"},
	{Std: "DD", Go: "02"},
	{Std: "HH", Go: "15"},
	{Std: "hh", Go: "03"},
	{Std: "h", Go: "3"},
	{Std: "mm", Go: "04"},
	{Std: "m", Go: "4"},
	{Std: "ss", Go: "05"},
	{Std: "s", Go: "5"},
}

// Format returns a textual representation of the time value formatted according
// to layout, which defines the format by showing how the reference time
// Example Format(time.Now(), "YYYY-MM-DD HH:mm:ss")
//
// layout defined by:
//  1. YYYY = 2006，YY = 06
//  2. MM = 01， MMM = Jan，MMMM = January
//  3. DD = 02，
//  4. DDD = Mon，DDDD = Monday
//  5. HH = 15，hh = 03, h = 3
//  6. mm = 04, m = 4
//  7. ss = 05, m = 5
func Format(t xtime.Time, layout string) string {
	for _, format := range formats {
		layout = strings.Replace(layout, format.Std, format.Go, 1)
	}

	return t.Format(layout)
}

func Parse(layout string, value string) (xtime.Time, error) {
	for _, format := range formats {
		layout = strings.Replace(layout, format.Std, format.Go, 1)
	}

	return xtime.Parse(layout, value)
}

func InRange(t xtime.Time, start, end xtime.Time) bool {
	return t.After(start) && t.Before(end)
}

func Random(max xtime.Duration) xtime.Duration {
	return xtime.Duration(rand.Int63n(int64(max)))
}

func Tick(ctx context.Context, d xtime.Duration, f func() error) error {
	ticker := xtime.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := f(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration xtime.Duration

func (d Duration) String() string {
	return xtime.Duration(d).String()
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case float64:
		*d = Duration(xtime.Duration(tv))

		return nil

	case string:
		parsed, err := xtime.ParseDuration(tv)
		if err != nil {
			return err
		}
		*d = Duration(parsed)

		return nil

	default:
		return fmt.Errorf("invalid duration: %v", tv)
	}
}

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := xtime.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := xtime.Until(deadline); ctimeout < xtime.Duration(d) {
			// deliver small timeout
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, xtime.Duration(d))
	return d, ctx, cancel
}

func TimeNewest(times ...xtime.Time) xtime.Time {
	newest := xtime.Time{}
	for _, t := range times {
		if t.After(newest) {
			newest = t
		}
	}
	return newest
}

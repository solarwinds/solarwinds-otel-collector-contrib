package swok8sdiscovery

import (
	"context"
	"time"
)

// Start ticking immediately.
// Ref: https://stackoverflow.com/questions/32705582/how-to-get-time-tick-to-tick-immediately
func newTicker(ctx context.Context, repeat time.Duration) *time.Ticker {
	ticker := time.NewTicker(repeat)
	oc := ticker.C
	nc := make(chan time.Time, 1)
	go func() {
		nc <- time.Now()
		for {
			select {
			case tm := <-oc:
				nc <- tm
			case <-ctx.Done():
				return
			}
		}
	}()

	ticker.C = nc
	return ticker
}

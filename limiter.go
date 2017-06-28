package rate_limiter

import (
	"context"
	"sync"
	"time"
)

const maxNextCallLagStreak = 10

// Used to rate-limit some resources, like network requests, disk access
type Limiter struct {
	sync.Mutex

	timePerReq time.Duration
	// closest time when next call is allowed
	nextCall       time.Time
	maxNextCallLag time.Duration
}

// Wait blocks to achieve average time between calls equals to timePerReq
// Could be interrupted using context, return error if interrupted
func (l *Limiter) Wait(ctx context.Context) (err error) {
	// check if context is already done
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	l.Lock()

	now := time.Now()

	// this is first call, allow it
	if l.nextCall.IsZero() {
		l.nextCall = now.Add(l.timePerReq)
		l.Unlock()

		return
	} else {
		var sleepTime time.Duration

		// Wait called before nextCall, we should sleep till nextCall
		if now.Before(l.nextCall) {
			sleepTime = l.nextCall.Sub(now)
		}

		// nextCall increment
		l.nextCall = l.nextCall.Add(l.timePerReq)

		// We shouldn't allow nextCall to be far behind from now
		// it would mean that service that slowed down for a short period of time
		// would get lots of requests following that
		if now.Sub(l.nextCall) > l.maxNextCallLag {
			l.nextCall = now.Add(-l.maxNextCallLag)
		}

		l.Unlock()

		if sleepTime > 0 {
			select {
			case <-ctx.Done():
				// context cancelled, decrease nextCall to reduce cancel influence
				l.Lock()
				l.nextCall = l.nextCall.Add(-l.timePerReq)
				l.Unlock()
				return ctx.Err()
			case <-time.After(sleepTime):
			}
		}
	}

	return
}

func NewLimiter(count int, per time.Duration) *Limiter {
	l := &Limiter{
		timePerReq: per / time.Duration(count),
	}
	l.maxNextCallLag = l.timePerReq * maxNextCallLagStreak

	return l
}

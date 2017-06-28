package rate_limiter

import (
	"sync/atomic"
	"testing"
	"time"

	"context"

	"sync"

	"math/rand"

	"fmt"

	"github.com/stretchr/testify/assert"
)

// test limiter with cancelled contexts, cancel shouldn't affect rate limit
func testLimiterWaitWithParams(t *testing.T, count int, per time.Duration, concurrentTasks int, testTime time.Duration, cancelContext bool) {
	l := NewLimiter(count, per)

	var initWg sync.WaitGroup
	var start = make(chan struct{})
	var done = make(chan struct{})
	var totalCount float64 = float64(testTime) / float64(per) * float64(count)
	var counter int64
	var increment = func() {
		atomic.AddInt64(&counter, 1)
	}

	var defaultContext = context.Background()

	for i := 0; i < concurrentTasks; i++ {
		initWg.Add(1)
		go func() {
			initWg.Done()
			// wait for start
			<-start

			for {
				select {
				case <-done:
				default:
				}

				var ctx context.Context

				// use cancel context param
				if cancelContext {
					// cancel each third request
					if rand.Float64() < (1.0 / 3.0) {
						// randomly calculate context cancel time from 0 to timePerReq
						deadlineTime := time.Duration(rand.Float64() * float64(l.timePerReq))
						ctx, _ = context.WithDeadline(defaultContext, time.Now().Add(deadlineTime))
					} else {
						ctx = defaultContext
					}
				} else {
					ctx = defaultContext
				}

				if err := l.Wait(ctx); err == nil {
					increment()
				}
			}
		}()
	}

	// wait all goroutines to start
	initWg.Wait()
	close(start)

	time.AfterFunc(testTime, func() {
		msg := fmt.Sprintf("count: %d, per: %v, tasks: %d, cancelContext: %v", count, per, concurrentTasks, cancelContext)
		assert.InDelta(t, totalCount, counter, maxNextCallLagStreak, msg)
		close(done)
	})
	<-done
}

func TestLimiterWait(t *testing.T) {
	type testParam struct {
		Count      int
		Per        time.Duration
		Concurrent int
	}
	pairs := []testParam{
		{1, time.Second, 10},
		{100, time.Second, 10},
		{1000, time.Second, 10},
		{10, time.Minute, 10},
		{1000, time.Minute, 10},
		{1000, 5 * time.Second, 10},
	}

	for i := 0; i < len(pairs); i++ {
		// test for both cancelContext true and false
		for j := 0; j < 2; j++ {
			var cancelContext bool
			if j == 0 {
				cancelContext = false
			} else {
				cancelContext = true
			}
			testLimiterWaitWithParams(t, pairs[i].Count, pairs[i].Per, pairs[i].Concurrent, time.Second*5, cancelContext)
		}
	}
}

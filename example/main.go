package main

import (
	"context"
	"sync"
	"time"

	"github.com/Hepri/rate_limiter"
)

func main() {
	rate := rate_limiter.NewLimiter(10, time.Second) // 10 rps

	for i := 0; i < 10; i++ {
		// wait until operation is permitted
		rate.Wait(context.Background())

		// do rate limited operation...
	}

	// concurrent example, multiple workers trying to use same resource with limited waiting time
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(200*time.Millisecond))
				if err := rate.Wait(ctx); err != nil {
					// wait is interrupted, handle this...
				}

				// do rate limited operation...
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

# Rate limiter
 
This library provides customisable rate-limiting algorithm using time.Sleep with ability to cancel waiting using context.

## Usage 

Create limiter with maximum count of operations per time unit. 
Before each operation you should call Wait(). Wait will sleep until you can continue.
You can cancel Wait anytime using context, in this case Wait will return error and request rate won't be affected. 

```go
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

```
package rate

import (
	"sync"
	"time"
)

type limiter struct {
	name       string
	capacity   int
	size       int
	lastAccess time.Time
	mu         *sync.Mutex
}

func NewLimiter(capacity int) {

}

package mask

import (
	"math/rand"
	"time"
)

func Jitter(min, max int) {
	time.Sleep(time.Duration(rand.Intn(max-min)+min) * time.Millisecond)
}

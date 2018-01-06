package gateway

import (
	"math/rand"
	"time"
)

type Batch []string

type Fake struct{}

func (f *Fake) Process(batch Batch) {
	if len(batch) == 0 {
		return
	}
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
}

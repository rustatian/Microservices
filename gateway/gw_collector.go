package gateway

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Processor interface {
	Process(Batch)
}

type Collector struct {
	Processor    Processor
	MaxBatchSize int
	WorkerCount  int
	QueueSize    int

	FlushInterval time.Duration
	payloadsQueue chan string
}

func (c *Collector) Collect(payload string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("collector is not running")
		}
	}()
	c.payloadsQueue <- payload
	log.Printf("collected: %s", payload)
	return nil
}

func (c *Collector) Run(stop chan struct{}) {
	log.Print("collector start")
	defer log.Println("collector stop")

	c.payloadsQueue = make(chan string, c.QueueSize)
	wg := sync.WaitGroup{}
	wg.Add(c.WorkerCount)

	for i := 0; i < c.WorkerCount; i++ {
		go func(i int) {
			defer wg.Done()
			c.worker(i)
		}(i)
	}
	<-stop
	close(c.payloadsQueue)
	wg.Wait()
}

func (c *Collector) worker(id int) {
	var buffer Batch
	log.Printf("worker_%d start", id)
	timer := time.NewTimer(c.FlushInterval)
	defer timer.Stop()

	for {
		select {
		case payload, opened := <-c.payloadsQueue:
			if !opened {
				c.flush(id, buffer, "stop")
				return
			}
			buffer = append(buffer, payload)
			if len(buffer) >= c.MaxBatchSize {
				c.flush(id, buffer, "size")
				buffer = nil
				timer.Reset(c.FlushInterval)
			}
		case <-timer.C:
			c.flush(id, buffer, "timer")
			buffer = nil
			timer.Reset(c.FlushInterval)
		}
	}
}

func (c *Collector) flush(workerId int, batch Batch, reason string) {
	t := time.Now()
	c.Processor.Process(batch)
	log.Printf("worker_%d flushed %d payloads by '%s' in %s", workerId, len(batch), reason, time.Since(t))
}

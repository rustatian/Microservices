package nats

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type Server struct {
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	before []ServerRequestFunc
	after  []ServerResponseFunc
	logger log.Logger

	FlushInterval time.Duration
	Conn          *nats.Conn
	MaxBatchSize  int
	WorkerCount   int
	QueueSize     int
	QueueCh       chan *nats.Msg
	ErrorCh       chan error
}

func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,

	MaxBatchSize int,
	WorkerCount int,
	QueueSize int,
	FlushInterval time.Duration,
	ErrorCh chan error,

	options ...ServerOption,

) *Server {
	conn, err := nats.Connect("nats://172.24.231.70:4222")
	if err != nil {
		panic(err)
	}
	s := &Server{
		e:      e,
		dec:    dec,
		enc:    enc,
		logger: *log.New(),

		Conn:          conn,
		FlushInterval: FlushInterval, //FlushInterval
		MaxBatchSize:  MaxBatchSize,
		WorkerCount:   WorkerCount,
		QueueSize:     QueueSize,
		QueueCh:       make(chan *nats.Msg), //QueueCh
		ErrorCh:       ErrorCh,
	}
	for _, option := range options {
		option(s)
	}

	go s.StartCollectLogs(nil)
	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

func ServerBefore(before ...ServerRequestFunc) ServerOption {
	return func(s *Server) {
		s.before = append(s.before, before...)
	}
}

func ServerAfter(after ...ServerResponseFunc) ServerOption {
	return func(s *Server) {
		s.after = append(s.after, after...)
	}
}

func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

func (s *Server) MsgHandler(msg *nats.Msg) {
	s.QueueCh <- msg
	log.Println("added in chan")
}

func (s *Server) worker(id int) {
	var buffer []*nats.Msg
	log.Printf("worker_%d start", id)
	timer := time.NewTimer(s.FlushInterval)
	defer timer.Stop()

	for {
		select {
		case payload, opened := <-s.QueueCh:
			if !opened {
				if len(buffer) == 0 {
					log.Printf("buffer is empty, get stop signal, worker %d", id)
					return
				}
				s.flush(id, buffer, "get stop signal")
				return
			}

			buffer = append(buffer, payload)
			if len(buffer) >= s.MaxBatchSize {
				s.flush(id, buffer, "max size reached")
				buffer = nil
				timer.Reset(s.FlushInterval)
			}
		case <-timer.C:
			//To prevent flushing empty buffer
			if len(buffer) == 0 {
				buffer = nil
				timer.Reset(s.FlushInterval)
			} else {
				s.flush(id, buffer, "time limit reached")
				buffer = nil
				timer.Reset(s.FlushInterval)
			}
		}
	}
}

func (s *Server) flush(workerId int, buffer []*nats.Msg, reason string) {

	defer func(tt time.Time) {
		log.Printf("worker_%d flushed %d payloads by '%s' in %s", workerId, len(buffer), reason, time.Since(tt))
	}(time.Now())

	for _, m := range buffer {
		// Non-nil non empty context to take the place of the first context in th chain of handling.
		ctx := context.TODO()

		request, err := s.dec(ctx, m)
		if err != nil {
			s.logger.Error("err", err)
			s.ErrorCh <- err
			return
		}

		response, err := s.e(ctx, request)
		if err != nil {
			s.logger.Error("err", err)
			s.ErrorCh <- err
			return
		}

		payload, err := s.enc(ctx, response)
		if err != nil {
			s.logger.Error("err", err)
			s.ErrorCh <- err
			return
		}

		s.Conn.Publish(m.Reply, payload)
	}
}

func (s *Server) StartCollectLogs(stop chan os.Signal) {
	log.Println("Start collect logs")
	defer log.Println("End collect logs")
	defer os.Exit(0)

	s.ErrorCh = make(chan error)
	s.QueueCh = make(chan *nats.Msg, s.QueueSize)
	//defer close(coll.ErrorCh)
	//defer close(coll.QueueCh)

	wg := sync.WaitGroup{}
	wg.Add(s.WorkerCount)
	for i := 0; i < s.WorkerCount; i++ {
		go func(id int) {
			defer wg.Done()
			s.worker(id)
		}(i)
	}
	<-stop
	close(s.QueueCh)
	wg.Wait()
	//close(coll.ErrorCh)
}

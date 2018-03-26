package old

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

var wg *sync.WaitGroup

type Server struct {
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	before []ServerRequestFunc
	after  []ServerResponseFunc
	logger logrus.Logger

	wg sync.WaitGroup

	Conn        *nats.Conn
	WorkerCount int
	QueueName   string
	StopCh      chan os.Signal
	QueueCh     chan *nats.Msg
	ErrorCh     chan error
}

func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,

	NatsConn *nats.Conn,
	Logger logrus.Logger,
	StopCh chan os.Signal,
	WorkerCount int,
	QueueName string,
	ErrorCh chan error,

	options ...ServerOption,

) *Server {
	s := &Server{
		e:      e,
		dec:    dec,
		enc:    enc,
		logger: Logger,

		wg: *wg,

		Conn:        NatsConn,
		WorkerCount: WorkerCount,
		QueueName:   QueueName,
		StopCh:      StopCh,
		QueueCh:     make(chan *nats.Msg), // QueueCh
		ErrorCh:     ErrorCh,
	}

	for _, option := range options {
		option(s)
	}

	go s.StartNatsWorkers(StopCh, NatsConn, QueueName)
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

func ServerErrorLogger(logger logrus.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

func (s *Server) MsgHandler(msg *nats.Msg) {
	s.QueueCh <- msg
	logrus.WithFields(
		logrus.Fields{
			"added": msg.Subject,
		},
	).Info("MsgHandler")
}

func (s *Server) worker(id int) {
	logrus.WithFields(
		logrus.Fields{
			"worker started, N: ": id,
		},
	).Info("MsgHandler")

	for {
		select {
		case payload, opened := <-s.QueueCh:
			if !opened {
				logrus.WithFields(
					logrus.Fields{
						"channel closed, worker N: ": id,
					},
				).Info("worker")

				return
			}
			s.flush(id, payload, "get stop signal")
		}
	}
}

func (s *Server) flush(workerId int, buffer *nats.Msg, reason string) {
	defer func(tt time.Time) {
		logrus.WithFields(
			logrus.Fields{
				"worker N ": workerId,
				"reason":    reason,
				"ts":        time.Since(tt),
			},
		).Info("flush")
	}(time.Now())

	ctx := context.TODO()
	request, err := s.dec(ctx, buffer)
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

	s.Conn.Publish(buffer.Reply, payload)
}

func (s *Server) StartNatsWorkers(stop chan os.Signal, nc *nats.Conn, queueName string) {
	logrus.WithFields(
		logrus.Fields{
			"start queue": queueName,
		},
	).Info("Start workers")

	defer logrus.WithFields(
		logrus.Fields{
			"end queue": queueName,
		},
	).Info("workers stopped")

	s.ErrorCh = make(chan error)

	s.wg.Add(s.WorkerCount)
	for i := 0; i < s.WorkerCount; i++ {
		go func(id int) {
			defer s.wg.Done()
			s.worker(id)
		}(i)
	}

	<-stop
	nc.Close()
	close(s.QueueCh)
	wg.Wait()
}

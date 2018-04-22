package nsq

import (
	"sync"
	"time"

	"github.com/go-mixins/queue"
	nsq "github.com/nsqio/go-nsq"
)

type mapKey struct {
	Topic, Channel string
}

// Subscriber implements queue.Subscriber with NSQ
type Subscriber struct {
	config
	consumers map[mapKey]*nsq.Consumer
	l         sync.RWMutex
}

// NewSubscriber creates the Subscriber
func NewSubscriber(opts ...Option) (res *Subscriber, err error) {
	res = new(Subscriber)
	res.config = config{options: nsq.NewConfig()}
	for _, opt := range opts {
		if err = opt(&res.config); err != nil {
			return
		}
	}
	if len(res.nsqlookupds) == 0 && len(res.nsqds) == 0 {
		res.nsqlookupds = append(res.nsqlookupds, "localhost:4161")
	}
	res.consumers = make(map[mapKey]*nsq.Consumer)
	return
}

// Connect starts message processing. Further Subscribes will panic.
func (s *Subscriber) Connect() (err error) {
	s.l.RLock()
	defer s.l.RUnlock()
	for _, c := range s.consumers {
		if len(s.nsqds) != 0 {
			err = c.ConnectToNSQDs(s.nsqds)
		} else {
			err = c.ConnectToNSQLookupds(s.nsqlookupds)
		}
		if err != nil {
			break
		}
	}
	return
}

func convert(handler queue.Handler) nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		err := handler(msg.Body)
		switch t := err.(type) {
		case queue.Delay:
			msg.RequeueWithoutBackoff(time.Duration(t))
			err = nil
		}
		return err
	}
}

// Subscribe creates new consumer for given topic and channel and assigns handler with optional concurency level
func (s *Subscriber) Subscribe(topic, channel string, handler queue.Handler, options ...queue.Option) (err error) {
	s.l.Lock()
	defer s.l.Unlock()
	key := mapKey{topic, channel}
	consumer := s.consumers[key]
	if consumer == nil {
		consumer, err = nsq.NewConsumer(topic, channel, s.options)
		if err != nil {
			return err
		}
		if s.logger != nil {
			consumer.SetLogger(s, s.level)
		}
		s.consumers[key] = consumer
	}
	c := 1
	for i := len(options) - 1; i >= 0; i-- {
		switch t := options[i].(type) {
		case queue.Concurrency:
			c = int(t)
		case queue.Middleware:
			handler = t(handler)
		}
	}
	consumer.AddConcurrentHandlers(convert(handler), c)
	consumer.ChangeMaxInFlight(c * 100) // XXX
	return
}

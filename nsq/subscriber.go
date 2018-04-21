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
	// NSQlookupD connection URL for consumers
	NSQLookupD string

	// List of NSQD addresses to connect if NSQLookupD is not specified
	NSQDs []string

	// NSQ options. Specified as a comma-separated list of key=value pairs.
	Options string

	// Log level
	LogLevel LogLevel

	// Filter out TOPIC_NOT_FOUND log messages by default
	KeepNsqLookupD404 bool

	// Logger for NSQ-specific messages
	Logger Logger

	consumers map[mapKey]*nsq.Consumer

	l sync.Mutex
}

// Connect starts message processing. Further Subscribes will panic.
func (s *Subscriber) Connect() (err error) {
	if s.NSQLookupD == "" && len(s.NSQDs) == 0 {
		s.NSQLookupD = "localhost:4161"
	}
	for _, c := range s.consumers {
		if len(s.NSQDs) != 0 {
			err = c.ConnectToNSQDs(s.NSQDs)
		} else {
			err = c.ConnectToNSQLookupd(s.NSQLookupD)
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
	if s.consumers == nil {
		s.consumers = make(map[mapKey]*nsq.Consumer)
	}
	consumer := s.consumers[key]
	if consumer == nil {
		config, err := toNSQConfig(s.Options)
		if err != nil {
			return err
		}
		consumer, err = nsq.NewConsumer(topic, channel, config)
		if err != nil {
			return err
		}
		if s.Logger != nil {
			consumer.SetLogger(logger{s.KeepNsqLookupD404, s.Logger}, s.LogLevel.toNSQ())
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

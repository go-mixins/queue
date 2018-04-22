package nsq

import (
	"strings"

	nsq "github.com/nsqio/go-nsq"
)

type config struct {
	nsqds       []string
	nsqlookupds []string
	marshal     func(interface{}) ([]byte, error)
	unmarshal   func([]byte) (interface{}, error)
	factories   map[string]func([]byte) (interface{}, error)
	options     *nsq.Config
	keep404     bool
	logger      Logger
	level       nsq.LogLevel
}

// Option modifies Publisher and Subscriber configuration
type Option func(*config) error

// TopicFactory specifies converter from wire format to Go object for the specified topic
func TopicFactory(topic string, f func([]byte) (interface{}, error)) Option {
	return func(dest *config) error {
		dest.factories[topic] = f
		return nil
	}
}

// Unmarshaler specifies default factory for topics
func Unmarshaler(f func([]byte) (interface{}, error)) Option {
	return func(dest *config) error {
		dest.unmarshal = f
		return nil
	}
}

// Marshaler specifies converter to wire format
func Marshaler(m func(interface{}) ([]byte, error)) Option {
	return func(dest *config) error {
		dest.marshal = m
		return nil
	}
}

// Queue adds NSQD instance URL(s) to configuration
func Queue(urls ...string) Option {
	return func(dest *config) error {
		dest.nsqds = append(dest.nsqds, urls...)
		return nil
	}
}

// Lookup adds NSQLookupD instance URL(s) configuration
func Lookup(urls ...string) Option {
	return func(dest *config) error {
		dest.nsqlookupds = append(dest.nsqlookupds, urls...)
		return nil
	}
}

// Keep404Errors ensures that TOPIC_NOT_FOUND errors are logged
var Keep404Errors Option = func(dest *config) error {
	dest.keep404 = true
	return nil
}

// Log configures logger for underlying NSQ producer or consumers
func Log(l Logger, level nsq.LogLevel) Option {
	return func(dest *config) error {
		dest.logger = l
		dest.level = level
		return nil
	}
}

// Options sets NSQ-specific options
func Options(opts ...string) Option {
	return func(dest *config) (err error) {
		for _, s := range opts {
			for _, opt := range strings.Split(s, ",") {
				vals := strings.SplitN(strings.TrimSpace(opt), "=", 2)
				var val interface{}
				if len(vals) == 1 {
					val = true
				} else {
					val = vals[1]
				}
				if err = dest.options.Set(vals[0], val); err != nil {
					return
				}
			}
		}
		return nil
	}
}

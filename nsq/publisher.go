// Package nsq implements Publisher and Subscriber using NSQ library and JSON
// serialization
package nsq

import (
	"encoding/json"
	"time"

	"github.com/nsqio/go-nsq"

	"github.com/go-mixins/queue"
)

// Publisher implements queue.Publisher with NSQ
type Publisher struct {
	// NSQD connection URL for producer
	NSQD string

	// NSQ options. Specified as a comma-separated list of key=value pairs.
	Options string

	// Log level
	LogLevel LogLevel

	// Logger for NSQ-specific messages
	Logger Logger

	// Converter to wire format
	Marshaler func(interface{}) ([]byte, error)

	producer *nsq.Producer
}

var _ queue.Publisher = (*Publisher)(nil)

// Open initializes the Publisher with sensible defaults and connects to NSQD
func (p *Publisher) Open() (err error) {
	if p.NSQD == "" {
		p.NSQD = "localhost:4150"
	}
	if p.Marshaler == nil {
		p.Marshaler = json.Marshal
	}
	cfg, err := toNSQConfig(p.Options)
	if err != nil {
		return
	}
	if p.producer, err = nsq.NewProducer(p.NSQD, cfg); err != nil {
		return
	}
	if p.Logger != nil {
		p.producer.SetLogger(logger{Logger: p.Logger}, p.LogLevel.toNSQ())
	}
	return p.producer.Ping()
}

// Publish marshals object to JSON and sends it to the specified topic
func (p *Publisher) Publish(topic string, obj interface{}, options ...queue.Option) (err error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return
	}
	pf := p.producer.Publish
	for _, opt := range options {
		switch t := opt.(type) {
		case queue.Delay:
			pf = func(topic string, jsonData []byte) error {
				return p.producer.DeferredPublish(topic, time.Duration(t), jsonData)
			}
		}
	}
	return pf(topic, jsonData)
}

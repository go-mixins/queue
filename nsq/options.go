package nsq

import (
	"strings"

	nsq "github.com/nsqio/go-nsq"
)

func toNSQConfig(s string) (res *nsq.Config, err error) {
	res = nsq.NewConfig()
	for _, opt := range strings.Split(s, ",") {
		vals := strings.SplitN(strings.TrimSpace(opt), "=", 2)
		var val interface{}
		if len(vals) == 1 {
			val = true
		} else {
			val = vals[1]
		}
		if err = res.Set(vals[0], val); err != nil {
			return
		}
	}
	return
}

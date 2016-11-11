package collector

import (
	rancher "github.com/rancher/go-rancher/v2"
	"github.com/rancher/telemetry/record"
)

type CollectorOpts struct {
	Client *rancher.RancherClient
}

type Collector interface {
	RecordKey() string
	Collect(opt *CollectorOpts) interface{}
}

var registered []Collector

func Register(c Collector) {
	registered = append(registered, c)
}

func Run(record *record.Record, opt *CollectorOpts) {
	for _, c := range registered {
		(*record)[c.RecordKey()] = c.Collect(opt)
	}
}

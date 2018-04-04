package collector

import (
	"github.com/rancher/telemetry/record"
	rancher "github.com/rancher/types/client/management/v3"
)

type CollectorOpts struct {
	Client *rancher.Client
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

package collector

import (
	"github.com/rancher/telemetry/record"
	rancherCluster "github.com/rancher/types/client/cluster/v3"
	rancher "github.com/rancher/types/client/management/v3"
	rancherProject "github.com/rancher/types/client/project/v3"
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

func GetClusterClient(c *CollectorOpts, id string) (*rancherCluster.Client, error) {
	options := *c.Client.Opts
	options.URL = options.URL + "/clusters/" + id

	return rancherCluster.NewClient(&options)
}

func GetProjectClient(c *CollectorOpts, id string) (*rancherProject.Client, error) {
	options := *c.Client.Opts
	options.URL = options.URL + "/projects/" + id

	return rancherProject.NewClient(&options)
}

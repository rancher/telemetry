package collector

import (
	rancherCluster "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/rancher/telemetry/record"
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

package collector

import (
	"time"

	"github.com/rancher/telemetry/record"
	rancherCluster "github.com/rancher/types/client/cluster/v3"
	rancher "github.com/rancher/types/client/management/v3"
	rancherProject "github.com/rancher/types/client/project/v3"
)

const (
	RECORD_INSTALLATION = "installation"
	RECORD_LICENSING    = "licensing"
	RECORD_VERSION      = 2
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

func Run(opt *CollectorOpts) map[string]record.Record {
	telemetryEnabled := IsTelemetryEnabled(opt)
	licensed := IsLicensed(opt)

	now := time.Now().UTC().Format(time.RFC3339)
	r := map[string]record.Record{}

	if telemetryEnabled {
		r[RECORD_INSTALLATION] = record.Record{
			"r":  RECORD_VERSION,
			"ts": now,
		}
	}

	if licensed {
		r[RECORD_LICENSING] = record.Record{
			"r":  RECORD_VERSION,
			"ts": now,
		}
	}

	for _, c := range registered {
		if c.RecordKey() == "license" && licensed {
			r[RECORD_LICENSING][c.RecordKey()] = c.Collect(opt)
			if !telemetryEnabled {
				return r
			}
			continue
		}
		if telemetryEnabled {
			r[RECORD_INSTALLATION][c.RecordKey()] = c.Collect(opt)
		}
	}

	return r
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

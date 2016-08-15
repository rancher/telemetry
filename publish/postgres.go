package publish

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	record "github.com/vincent99/telemetry/record"
)

type Postgres struct {
	telemetryVersion string
	host             string
	port             string
	user             string
	pass             string
}

func NewPostgres(c *cli.Context) *Postgres {
	out := &Postgres{
		telemetryVersion: c.App.Version,
		host:             c.String("pg-host"),
		port:             c.String("pg-port"),
		user:             c.String("pg-user"),
		pass:             c.String("pg-pass"),
	}

	if out.host == "" || out.user == "" || out.pass == "" {
		log.Warn("pg-{host,user,pass} options are required to publish to Postgres")
	}

	return out
}

func (p *Postgres) Report(r record.Record, clientIp string) error {
	log.Debugf("Publishing to Postgres")
	return nil
}

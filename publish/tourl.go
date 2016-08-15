package publish

import (
	"bytes"
	"encoding/json"

	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	record "github.com/vincent99/telemetry/record"
)

type ToUrl struct {
	url              string
	uid              string
	telemetryVersion string
	rancherImage     string
	rancherVersion   string
}

func NewToUrl(c *cli.Context) *ToUrl {
	out := &ToUrl{
		telemetryVersion: c.App.Version,
		url:              c.String("to-url"),
	}

	if out.url == "" {
		log.Warn("No to-url configured, not publishing")
	}

	return out
}

func (p *ToUrl) Report(r record.Record, clientIp string) error {
	if p.url == "" {
		return nil
	}

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	res, err := http.Post(p.url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	return err
}

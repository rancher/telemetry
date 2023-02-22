package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	record "github.com/rancher/telemetry/record"
)

type ToUrl struct {
	url              string
	telemetryVersion string
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
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		log.Debugf(fmt.Sprintf("Server said %d: %s", res.StatusCode, body))
		return nil
	} else {
		log.Errorf(fmt.Sprintf("Server said %d: %s", res.StatusCode, body))
		return fmt.Errorf("Server returned %d", res.StatusCode)
	}
}

package publish

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	collector "github.com/rancher/telemetry/collector"
	record "github.com/rancher/telemetry/record"
)

const GA_URL = "https://www.google-analytics.com/collect"

type Google struct {
	tid              string
	telemetryVersion string
}

type GoogleOpts struct {
	tid              string
	telemetryVersion string

	rancherImage   string
	rancherVersion string
	uid            string
	clientIp       string
}

func NewGoogle(c *cli.Context) *Google {
	out := &Google{
		telemetryVersion: c.App.Version,
		tid:              c.String("ga-tid"),
	}

	if out.tid != "" {
		log.Info("Google Analytics enabled")
	}

	return out
}

func (g *Google) Report(r record.Record, clientIp string) error {
	if g.tid == "" {
		return nil
	}

	log.Debugf("Publishing to Google Analytics")

	opts := &GoogleOpts{
		tid:              g.tid,
		telemetryVersion: g.telemetryVersion,
		clientIp:         clientIp,
	}

	// Ugly, but effective enough
	switch install := r["install"].(type) {
	case map[string]interface{}:
		opts.uid = install["uid"].(string)
		opts.rancherImage = install["image"].(string)
		opts.rancherVersion = install["version"].(string)
	}

	for category, entry := range r {
		g.flatten(opts, category, "", "", entry)
	}

	return nil
}

func (g *Google) flatten(opts *GoogleOpts, category string, event string, label string, vv interface{}) {
	//log.Debugf("Flattening %s/%s/%s = %s", category, event, label, vv)

	switch v := vv.(type) {

	case float64:
		g.sendEvent(opts, category, event, label, int64(collector.Round(v)))

	case int64:
		g.sendEvent(opts, category, event, label, v)

	case string:
		// Ignore

	case map[string]interface{}:
		for key, val := range v {
			if event == "" {
				g.flatten(opts, category, key, "", val)
			} else if label == "" {
				g.flatten(opts, category, event, key, val)
			} else {
				log.Errorf("Map too many levels deep for GA: %s/%s/%s -> %s", category, event, label, key)
			}
		}

	default:
		log.Warnf("Unknown flatten field input: %s/%s/%s: %s", category, event, label, v)
	}
}

func (g *Google) sendEvent(opts *GoogleOpts, category, action, label string, value int64) error {
	// Action is required, so ignore top-level keys like r=1
	if action == "" {
		return nil
	}

	qp := url.Values{}
	qp.Add("v", "1")                       // Protocol version
	qp.Add("tid", g.tid)                   // Tracking Account ID
	qp.Add("t", "event")                   // Hit type
	qp.Add("aip", "1")                     // Anonymize source IP addresses
	qp.Add("an", opts.rancherImage)        // App ID (Docker image)
	qp.Add("av", opts.rancherVersion)      // App Version (Image tag)
	qp.Add("aiid", opts.telemetryVersion)  // App ID (Docker image)
	qp.Add("cid", opts.uid)                // Installation UID
	qp.Add("ec", category)                 // Category
	qp.Add("ea", action)                   // Action
	qp.Add("ev", strconv.Itoa(int(value))) // Value

	// Client IP
	if opts.clientIp != "::1" && opts.clientIp != "127.0.0.1" {
		qp.Add("uip", opts.clientIp)
	}

	// Label
	if label != "" {
		qp.Add("el", label)
	}

	log.Debugf("Sending: %s", qp.Encode())

	res, err := http.PostForm(GA_URL, qp)
	if err != nil {
		log.Errorf("Error sending event: %s", err)
		return err
	}

	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	return err
}

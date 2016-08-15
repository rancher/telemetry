package publish

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	collector "github.com/vincent99/telemetry/collector"
	record "github.com/vincent99/telemetry/record"
)

const GA_URL = "https://www.google-analytics.com/collect"

type Google struct {
	tid              string
	uid              string
	telemetryVersion string
	rancherImage     string
	rancherVersion   string
}

func (g *Google) Configure(version string, c *cli.Context) {
	g.tid = c.String("ga-tid")
	if version == "" {
		g.telemetryVersion = "dev"
	} else {
		g.telemetryVersion = version
	}

	if g.tid == "" {
		log.Error("ga-tid option is required to publish to Google Analytics")
	}
}

func (g *Google) Report(r record.Record) {
	log.Debugf("Publishing to Google Analytics")

	// Ugly, but effective enough
	switch install := r["install"].(type) {
	case collector.Installation:
		g.uid = install.Uid
		g.rancherImage = install.Image
		g.rancherVersion = install.Version
	}

	for category, entry := range r {
		val := reflect.ValueOf(entry)
		if val.Kind() == reflect.Struct || val.Kind() == reflect.Map {
			g.flatten(category, "", val)
		}
	}
}

func (g *Google) flatten(category string, event string, v reflect.Value) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := t.Field(i).Name

		switch f.Kind() {
		case reflect.Int:
			if event == "" {
				g.sendEvent(category, name, "", f.Int())
			} else {
				g.sendEvent(category, event, name, f.Int())
			}

		case reflect.Struct:
			if event == "" {
				g.flatten(category, name, f)
			} else {
				log.Errorf("Struct too many levels deep for GA: %s/%s/%s", category, event, name)
			}

		case reflect.Map:
			if event == "" {
				iface := f.Interface()
				switch m := iface.(type) {
				case collector.LabelCount:
					for label, val := range m {
						g.sendEvent(category, name, label, int64(val))
					}
				default:
					log.Errorf("Unknown map type: %s\n", m)
				}
			} else {
				log.Errorf("Map too many levels deep for GA: %s/%s/%s", category, event, name)
			}

		case reflect.String:
			// Ignore

		default:
			log.Warnf("Unknown flatten field input: %s=%s", name, f)
		}
	}
}

func (g *Google) sendEvent(category, action, label string, value int64) error {
	qp := url.Values{}
	qp.Add("v", "1")                   // Protocol version
	qp.Add("tid", g.tid)               // Tracking Account ID
	qp.Add("t", "event")               // Hit type
	qp.Add("aip", "1")                 // Anonymize source IP addresses
	qp.Add("an", g.rancherImage)       // App ID (Docker image)
	qp.Add("av", g.rancherVersion)     // App Version (Image tag)
	qp.Add("aiid", g.telemetryVersion) // App ID (Docker image)
	qp.Add("cid", g.uid)               // Installation UID
	qp.Add("ec", category)             // Category
	qp.Add("ea", action)               // Action
	if label != "" {
		qp.Add("el", label) // Label
	}
	qp.Add("ev", strconv.Itoa(int(value))) // Value

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

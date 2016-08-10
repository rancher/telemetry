package collector

import (
	log "github.com/Sirupsen/logrus"
	"regexp"
)

type Installation struct {
	Image      string `json:"image"`
	Version    string `json:"version"`
	AuthConfig string `json:"auth"`
}

func (i Installation) RecordKey() string {
	return "install"
}

func (i Installation) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting Installation")

	if image, ok := GetSetting(c.Client, "rancher.server.image"); ok {
		log.Debugf("  Image: %s", image)
		i.Image = image
	}

	if version, ok := GetSetting(c.Client, "rancher.server.version"); ok {
		log.Debugf("  Version: %s", version)
		i.Version = version
	}

	// @TODO replace with unified authConfig
	if enabled, ok := GetSetting(c.Client, "api.security.enabled"); ok {
		if provider, ok := GetSetting(c.Client, "api.auth.provider.configured"); ok {
			if enabled == "true" {
				i.AuthConfig = regexp.MustCompile("(?i)^(.*?)config$").ReplaceAllString(provider, "$1")
			} else {
				i.AuthConfig = "none"
			}
		}
	}

	return i
}

/*
func settingsAsMap(client *rancher.RancherClient) (map[string]string, error) {
	list, err := c.Client.Setting.List(&rancher.ListOpts{})

	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	for _, s := range list {
		out[s.Name] = s.Value
	}

	return out
}
*/

func init() {
	Register(Installation{})
}

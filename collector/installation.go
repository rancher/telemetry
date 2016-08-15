package collector

import (
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"regexp"
)

const UID_SETTING = "telemetry.uid"

type Installation struct {
	Uid        string     `json:"uid"`
	Image      string     `json:"image"`
	Version    string     `json:"version"`
	AuthConfig LabelCount `json:"auth"`
}

func (i Installation) RecordKey() string {
	return "install"
}

func (i Installation) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting Installation")

	i.Uid = i.GetUid(c)
	i.Image = "unknown"
	i.Version = "unknown"
	i.AuthConfig = make(LabelCount)

	if image, ok := GetSetting(c.Client, "rancher.server.image"); ok {
		log.Debugf("  Image: %s", image)
		if image != "" {
			i.Image = image
		}
	}

	if version, ok := GetSetting(c.Client, "rancher.server.version"); ok {
		log.Debugf("  Version: %s", version)
		if version != "" {
			i.Version = version
		}
	}

	// @TODO replace with unified authConfig
	authConfig := "none"
	if enabled, ok := GetSetting(c.Client, "api.security.enabled"); ok {
		if provider, ok := GetSetting(c.Client, "api.auth.provider.configured"); ok {
			if enabled == "true" {
				authConfig = regexp.MustCompile("(?i)^(.*?)config$").ReplaceAllString(provider, "$1")
			}
		}
	}
	i.AuthConfig.Increment(authConfig)

	return i
}

func (i Installation) GetUid(c *CollectorOpts) string {
	uid, ok := GetSetting(c.Client, UID_SETTING)
	if ok && uid != "" {
		log.Debugf("  Using Existing Uid: %s", uid)
		return uid
	}

	uid = uuid.NewV4().String()
	err := SetSetting(c.Client, UID_SETTING, uid)

	if err == nil {
		log.Debugf("  Generated Uid: %s", uid)
		return uid
	} else {
		log.Debugf("  Error Generating Uid: %s", err)
		return ""
	}
}

func init() {
	Register(Installation{})
}

package collector

import (
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	UID_SETTING            = "install-uuid"
	SERVER_IMAGE_SETTING   = "server-image"
	SERVER_VERSION_SETTING = "server-version"
)

type Installation struct {
	Uid     string `json:"uid"`
	Image   string `json:"image"`
	Version string `json:"version"`
	//AuthConfig LabelCount `json:"auth"`
}

func (i Installation) RecordKey() string {
	return "install"
}

func (i Installation) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting Installation")

	settings := GetSettingCollection(c.Client)

	uid, _ := GetSettingByCollection(settings, UID_SETTING)
	uid, _ = i.GetUid(uid, c)

	i.Uid = uid
	i.Image = "unknown"
	i.Version = "unknown"
	//i.AuthConfig = make(LabelCount)

	if image, ok := GetSettingByCollection(settings, SERVER_IMAGE_SETTING); ok {
		log.Debugf("  Image: %s", image)
		if image != "" {
			i.Image = image
		}
	}

	if version, ok := GetSettingByCollection(settings, SERVER_VERSION_SETTING); ok {
		log.Debugf("  Version: %s", version)
		if version != "" {
			i.Version = version
		}
	}

	// @TODO replace with unified authConfig
	/*authConfig := "none"
	if enabled, ok := GetSetting(c.Client, "api.security.enabled"); ok {
		if provider, ok := GetSetting(c.Client, "api.auth.provider.configured"); ok {
			if enabled == "true" {
				authConfig = regexp.MustCompile("(?i)^(.*?)config$").ReplaceAllString(provider, "$1")
			}
		}
	}
	i.AuthConfig.Increment(authConfig)
	*/

	return i
}

func (i Installation) GetUid(uid string, c *CollectorOpts) (string, bool) {
	if uid != "" {
		log.Debugf("  Using Existing Uid: %s", uid)
		return uid, true
	}

	newuid, _ := uuid.NewV4()
	uid = newuid.String()
	err := SetSetting(c.Client, UID_SETTING, uid)
	if err != nil {
		log.Debugf("  Error Generating Uid: %s", err)
		return "", false
	}

	log.Debugf("  Generated Uid: %s", uid)
	return uid, true
}

func init() {
	Register(Installation{})
}

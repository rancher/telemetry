package collector

import (
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	TELEMETRY_UID_SETTING      = "telemetry-uid"
	SERVER_VERSION_SETTING     = "server-version"
	UI_DEFAULT_LANDING_SETTING = "ui-default-landing"
	UI_LANDING_EXPLORER        = "vue"
	UI_LANDING_MANAGER         = "ember"
	UI_LANDING_DEFAULT         = UI_LANDING_MANAGER
)

type Installation struct {
	Uid                  string     `json:"uid"`
	Version              string     `json:"version"`
	UiLanding            string     `json:"uiLanding"`
	AuthConfig           LabelCount `json:"auth"`
	Users                LabelCount `json:"users"`
	KontainerDriverCount int        `json:"kontainerDriverCount"`
	KontainerDrivers     LabelCount `json:"kontainerDrivers"`
	NodeDriverCount      int        `json:"nodeDriverCount"`
	NodeDrivers          LabelCount `json:"nodeDrivers"`
	HasInternal          bool       `json:"hasInternal"`
}

func (i Installation) RecordKey() string {
	return "install"
}

func (i Installation) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting Installation")

	nonRemoved := NonRemoved()

	i.GetUid(c)
	i.GetVersion(c)
	i.GetUILanding(c)
	i.AuthConfig = make(LabelCount)
	i.Users = make(LabelCount)
	i.KontainerDrivers = make(LabelCount)
	i.NodeDrivers = make(LabelCount)

	log.Debug("  Collecting AuthConfigs")
	configList, err := c.Client.AuthConfig.ListAll(&nonRemoved)
	if err == nil {
		for _, config := range configList.Data {
			if config.Enabled {
				name := regexp.MustCompile("(?i)^(.*?)Config$").ReplaceAllString(config.Type, "$1")
				i.AuthConfig.Increment(name)
			}
		}
	} else {
		log.Errorf("Failed to get authProviders err=%s", err)
	}

	log.Debug("  Collecting Users")
	userList, err := c.Client.User.ListAll(&nonRemoved)
	if err == nil {
		for _, user := range userList.Data {
			for _, principalID := range user.PrincipalIDs {
				provider := strings.Split(principalID, "://")
				if len(provider) > 1 {
					i.Users.Increment(provider[0])
				}
			}
		}
	} else {
		log.Errorf("Failed to get users err=%s", err)
	}

	log.Debug("  Collecting NodeDrivers")
	nodeDriverList, err := c.Client.NodeDriver.ListAll(&nonRemoved)
	if err == nil {
		for _, driver := range nodeDriverList.Data {
			if driver.Active {
				i.NodeDrivers.Increment(driver.Name)
				i.NodeDriverCount++
			}
		}
	} else {
		log.Errorf("Failed to get nodeDrivers err=%s", err)
	}

	log.Debug("  Collecting KontainerDrivers")
	kontainerDriverList, err := c.Client.KontainerDriver.ListAll(&nonRemoved)
	if err == nil {
		for _, driver := range kontainerDriverList.Data {
			if driver.Active {
				i.KontainerDrivers.Increment(driver.Name)
				i.KontainerDriverCount++
			}
		}
	} else {
		log.Errorf("Failed to get kontainerDrivers err=%s", err)
	}

	i.HasInternal = false

	log.Debug("  Looking for Local cluser")
	clusterList, err := c.Client.Cluster.ListAll(&nonRemoved)
	if err == nil {
		for _, cluster := range clusterList.Data {
			if cluster.Internal {
				i.HasInternal = true
				break
			}
		}
	} else {
		log.Errorf("Failed to get Clusters err=%s", err)
	}

	return i
}

func (i *Installation) GetUILanding(c *CollectorOpts) {
	uiLanding, err := c.Client.Setting.ByID(UI_DEFAULT_LANDING_SETTING)
	if err != nil {
		if !IsNotFound(err) {
			log.Errorf("Failed to get setting %s err=%s", UI_DEFAULT_LANDING_SETTING, err)
		}
	}
	defer log.Debugf("  Installation UI Landing: %s", i.UiLanding)
	if uiLanding == nil || len(uiLanding.Value) == 0 {
		i.UiLanding = UI_LANDING_DEFAULT
		return
	}
	i.UiLanding = uiLanding.Value
}

func (i *Installation) GetVersion(c *CollectorOpts) {
	version, err := c.Client.Setting.ByID(SERVER_VERSION_SETTING)
	if err != nil {
		log.Errorf("Failed to get setting %s err=%s", SERVER_VERSION_SETTING, err)
	}
	defer log.Debugf("  Installation Server Version: %s", i.Version)
	if version == nil || len(version.Value) == 0 {
		i.Version = "unknown"
		return
	}

	i.Version = version.Value
}

func GetTelemetryUid(c *CollectorOpts) (string, bool) {
	telemetryUid, err := c.Client.Setting.ByID(TELEMETRY_UID_SETTING)
	if err != nil {
		if !IsNotFound(err) {
			log.Errorf("Failed to get setting %s err=%s", TELEMETRY_UID_SETTING, err)
			return "", false
		}
	}

	uid := telemetryUid.Value
	if len(uid) > 0 {
		log.Debugf("  Using Existing Telemetry Uid: %s", uid)
		return uid, true
	}

	newuid, _ := uuid.NewV4()
	uid = newuid.String()
	err = SetSetting(c.Client, TELEMETRY_UID_SETTING, uid)
	if err != nil {
		log.Errorf("Error Setting generated Telemetry Uid: %s", err)
		return "", false
	}

	log.Debugf("  Generated Telemetry Uid: %s", uid)
	return uid, true
}

func (i *Installation) GetUid(c *CollectorOpts) {
	i.Uid, _ = GetTelemetryUid(c)
}

func init() {
	Register(Installation{})
}

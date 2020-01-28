package collector

import (
	log "github.com/sirupsen/logrus"
)

const (
	SERVER_LICENSE_SETTING   = "server-license"
	INSTALLATION_UID_SETTING = "install-uuid"
)

type License struct {
	Key             string `json:"key"`
	InstallationUID string `json:"installationUid"`
	TelemetryUID    string `json:"telemetryUid"`
	RunningNodes    int    `json:"runningNodes"`
}

func (l License) RecordKey() string {
	return "license"
}

func (l License) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting License")

	log.Debug("  Getting key")
	if ok := l.GetKey(c); !ok {
		return nil
	}

	log.Debug("  Getting Installation uid")
	if ok := l.GetInstallationUid(c); !ok {
		return nil
	}

	l.TelemetryUID = "disabled"
	if IsTelemetryEnabled(c) {
		log.Debug("  Getting Telemetry uid")
		l.TelemetryUID, _ = GetTelemetryUid(c)
	}

	log.Debug("  Getting Local cluster")
	localClusterID := ""
	localFilter := NonRemoved()
	localFilter.Filters["internal"] = true
	clusterList, err := c.Client.Cluster.ListAll(&localFilter)
	if err != nil {
		log.Errorf("    Failed to get Local Cluster err=%s", err)
	}
	if clusterList != nil && len(clusterList.Data) == 1 {
		localClusterID = clusterList.Data[0].ID
		log.Debug("    Local cluster found")
	} else {
		log.Debug("    Local cluster NOT found")
	}

	log.Debug("  Getting Nodes")

	nodeFilter := NonRemoved()
	nodeFilter.Filters["clusterId_ne"] = localClusterID
	nodeList, err := c.Client.Node.ListAll(&nodeFilter)
	if err != nil {
		log.Errorf("    Failed collect running nodes info err=%s", err)
		return nil
	}
	log.Debugf("    Found %d Nodes", len(nodeList.Data))
	l.RunningNodes = len(nodeList.Data)

	return l
}

func GetLicenseKey(c *CollectorOpts) (string, bool) {
	licenseKey, err := c.Client.Setting.ByID(SERVER_LICENSE_SETTING)
	if err != nil {
		if IsNotFound(err) {
			log.Debugf("  Setting %s doesn't exist", SERVER_LICENSE_SETTING)
		} else {
			log.Errorf("  Failed to get setting %s err=%s", SERVER_LICENSE_SETTING, err)
		}
		return "", false
	}

	key := licenseKey.Value
	if len(key) == 0 {
		log.Debug("  No license key")
		return "", false
	}

	log.Debugf("  License key: %s", key)

	return key, true
}

func (l *License) GetKey(c *CollectorOpts) bool {
	key, ok := GetLicenseKey(c)
	l.Key = key

	return ok
}

func IsLicensed(c *CollectorOpts) bool {
	_, license := GetLicenseKey(c)
	return license
}

func GetInstallationUid(c *CollectorOpts) (string, bool) {
	installUid, err := c.Client.Setting.ByID(INSTALLATION_UID_SETTING)
	if err != nil {
		log.Errorf("  Failed to get setting %s err=%s", INSTALLATION_UID_SETTING, err)
		return "", false
	}

	return installUid.Value, true
}

func (l *License) GetInstallationUid(c *CollectorOpts) bool {
	installUid, ok := GetInstallationUid(c)
	l.InstallationUID = installUid

	return ok
}

func init() {
	Register(License{})
}

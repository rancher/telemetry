package collector

import (
	log "github.com/sirupsen/logrus"
)

type MultiClusterApp struct {
	Total        int     `json:"total"`
	Active       int     `json:"active"`
	TargetMin    int     `json:"targetMin"`
	TargetMax    int     `json:"targetMax"`
	TargetAvg    float64 `json:"targetAvg"`
	TargetTotal  int     `json:"targetTotal"`
	DnsProviders int     `json:"dnsProviders"`
	DnsEntries   int     `json:"dnsEntries"`
}

func (mca MultiClusterApp) RecordKey() string {
	return "mca"
}

func (mca MultiClusterApp) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Apps")
	appList, err := c.Client.MultiClusterApp.List(&nonRemoved)
	if err == nil {
		log.Debugf("  Found %d Apps", len(appList.Data))

		var targetCounts []float64

		// Clusters
		for _, app := range appList.Data {
			mca.Total++
			if app.State == "active" {
				mca.Active++
			}

			targets := len(app.Targets)

			mca.TargetTotal += targets
			mca.TargetMin = MinButNotZero(mca.TargetMin, targets)
			mca.TargetMax = Max(mca.TargetMax, targets)
			targetCounts = append(targetCounts, float64(targets))
		}

		mca.TargetAvg = Average(targetCounts)
	} else {
		log.Errorf("Failed to get Apps err=%s", err)
		return nil
	}

	// Global DNS Providers
	log.Debug("Collecting DNS Providers")
	dnsList, err := c.Client.GlobalDNSProvider.List(&nonRemoved)
	if err == nil {
		count := len(dnsList.Data)
		log.Debugf("  Found %d DNS Providers", count)
		mca.DnsProviders = count
	} else {
		log.Errorf("Failed to get DNS Providers err=%s", err)
		return nil
	}

	// Global DNS Entries
	log.Debug("Collecting DNS Entries")
	entryList, err := c.Client.GlobalDNS.List(&nonRemoved)
	if err == nil {
		count := len(entryList.Data)
		log.Debugf("  Found %d DNS Entries", count)
		mca.DnsEntries = count
	} else {
		log.Errorf("Failed to get DNS Entries err=%s", err)
		return nil
	}

	return mca
}

func init() {
	Register(MultiClusterApp{})
}

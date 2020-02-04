package collector

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type MultiClusterApp struct {
	Total        int                     `json:"total"`
	Active       int                     `json:"active"`
	TargetMin    int                     `json:"targetMin"`
	TargetMax    int                     `json:"targetMax"`
	TargetAvg    float64                 `json:"targetAvg"`
	TargetTotal  int                     `json:"targetTotal"`
	DnsProviders int                     `json:"dnsProviders"`
	DnsEntries   int                     `json:"dnsEntries"`
	Catalogs     map[string]*AppTemplate `json:"rancheCatalogs"`
}

func (mca MultiClusterApp) RecordKey() string {
	return "mca"
}

func (mca MultiClusterApp) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting MultiClusterApps")
	appList, err := c.Client.MultiClusterApp.ListAll(&nonRemoved)
	if err == nil {
		log.Debugf("  Found %d MultiClusterApps", len(appList.Data))

		var targetCounts []float64
		mca.Catalogs = map[string]*AppTemplate{}

		for _, catalog := range AppRancherCatalogs {
			state, err := GetAppCatalogState(c, catalog)
			if err != nil {
				log.Errorf("Failed to get Catalog ID %s err=%s", catalog, err)
				return nil
			}
			mca.Catalogs[catalog] = &AppTemplate{
				State: state,
				Apps:  map[string]*LabelCount{},
			}
		}

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

			templateVersion, err := c.Client.TemplateVersion.ByID(app.TemplateVersionID)
			if err != nil {
				continue
			}
			externalID, err := SplitMultiClusterAppExternalID(templateVersion.ExternalID)
			if err != nil {
				log.Errorf("Failed to split App External ID %s err=%s", templateVersion.ExternalID, err)
				continue
			}
			if mca.Catalogs[externalID["catalog"]] == nil {
				continue
			}

			if mca.Catalogs[externalID["catalog"]].Apps[externalID["template"]] == nil {
				mca.Catalogs[externalID["catalog"]].Apps[externalID["template"]] = &LabelCount{}
			}
			mca.Catalogs[externalID["catalog"]].Apps[externalID["template"]].Increment(externalID["version"])
		}

		mca.TargetAvg = Average(targetCounts)
	} else {
		log.Errorf("Failed to get Apps err=%s", err)
	}

	// Global DNS Providers (only with management cluster, so ignore errors)
	log.Debug("  Collecting DNS Providers")
	dnsList, err := c.Client.GlobalDNSProvider.ListAll(&nonRemoved)
	if err == nil {
		count := len(dnsList.Data)
		log.Debugf("    Found %d DNS Providers", count)
		mca.DnsProviders = count
	}

	// Global DNS Entries (only with management cluster, so ignore errors)
	log.Debug("  Collecting DNS Entries")
	entryList, err := c.Client.GlobalDNS.ListAll(&nonRemoved)
	if err == nil {
		count := len(entryList.Data)
		log.Debugf("    Found %d DNS Entries", count)
		mca.DnsEntries = count
	}

	return mca
}

func SplitMultiClusterAppExternalID(externalID string) (map[string]string, error) {
	//Global catalog url: catalog://?catalog=demo&template=test&version=1.23.0

	val, err := url.Parse(externalID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	out["catalog"] = val.Query().Get("catalog")
	out["template"] = val.Query().Get("template")
	out["version"] = val.Query().Get("version")

	if out["catalog"] == "" || out["template"] == "" || out["version"] == "" {
		return nil, fmt.Errorf("Bad External ID format")
	}

	return out, nil
}

func init() {
	Register(MultiClusterApp{})
}

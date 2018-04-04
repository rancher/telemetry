package collector

import (
	norman "github.com/rancher/norman/types"
	rancher "github.com/rancher/types/client/project/v3"
	log "github.com/sirupsen/logrus"
)

type WorkloadInfo struct {
	WorkloadMin   int `json:"min"`
	WorkloadMax   int `json:"max"`
	WorkloadTotal int `json:"total"`
	WorkloadAvg   int `json:"avg"`
}

func (w *WorkloadInfo) Update(i int) {
	w.WorkloadTotal += i
	w.WorkloadMin = MinButNotZero(w.WorkloadMin, i)
	w.WorkloadMax = Max(w.WorkloadMax, i)
}

func (w *WorkloadInfo) UpdateAvg(i []float64) {
	w.WorkloadAvg = Clamp(0, Round(Average(i)), 100)
}

func GetWorkloadCollection(c *CollectorOpts, url string) *rancher.WorkloadCollection {
	if url == "" {
		log.Debugf("Workload collection link is empty.")
		return nil
	}

	wlCollection := &rancher.WorkloadCollection{}
	version := "workloads"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, wlCollection)
	if err != nil {
		log.Debugf("Error getting workload collection [%s] %s", resource.Links[version], err)
		return nil
	}

	if wlCollection == nil || wlCollection.Type != "collection" || len(wlCollection.Data) == 0 {
		log.Debugf("Workload collection is empty [%s]", resource.Links[version])
		return nil
	}

	return wlCollection
}

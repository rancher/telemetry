package collector

import (
	log "github.com/sirupsen/logrus"
)

const orchestrationName = "cattle-V2.0"

type Project struct {
	Total         int          `json:"total"`
	Ns            NsInfo       `json:"namespace"`
	Workload      WorkloadInfo `json:"workload"`
	Pod           PodData      `json:"pod"`
	Orchestration LabelCount   `json:"orch"`
}

func (p Project) RecordKey() string {
	return "project"
}

func (p Project) Collect(c *CollectorOpts) interface{} {
	opts := NonRemoved()
	opts.Filters["all"] = "true"

	log.Debug("Collecting Projects")
	list, err := c.Client.Project.List(&opts)

	if err != nil {
		log.Errorf("Failed to get Projects err=%s", err)
		return nil
	}

	total := len(list.Data)
	log.Debugf("  Found %d Projects", total)

	p.Orchestration = make(LabelCount)
	p.Orchestration[orchestrationName] = total
	p.Total = total

	var nsUtils []float64
	var wlUtils []float64
	var poUtils []float64

	for _, project := range list.Data {
		// Namespace
		nsCollection := GetNamespaceCollection(c, project.Links["namespaces"])
		if nsCollection != nil {
			totalNs := len(nsCollection.Data)
			p.Ns.Update(totalNs)
			nsUtils = append(nsUtils, float64(totalNs))
			p.Ns.UpdateDetails(nsCollection)
		}

		// Workload
		wlCollection := GetWorkloadCollection(c, project.Links["workloads"])
		if wlCollection != nil {
			totalWl := len(wlCollection.Data)
			p.Workload.Update(totalWl)
			wlUtils = append(wlUtils, float64(totalWl))
		}

		// Pod
		poCollection := GetPodCollection(c, project.Links["pods"])
		if poCollection != nil {
			totalPo := len(poCollection.Data)
			p.Pod.Update(totalPo)
			poUtils = append(poUtils, float64(totalPo))
		}
	}

	p.Ns.UpdateAvg(nsUtils)
	p.Workload.UpdateAvg(wlUtils)
	p.Pod.UpdateAvg(poUtils)

	return p
}

func init() {
	Register(Project{})
}

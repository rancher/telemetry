package collector

import (
	"strings"

	rancher "github.com/rancher/types/client/cluster/v3"
	log "github.com/sirupsen/logrus"
)

const orchestrationName = "cattle-V2.0"

const projectLabel = "field.cattle.io/projectId"

type Project struct {
	Total         int          `json:"total"`
	Ns            NsInfo       `json:"namespace"`
	Workload      WorkloadInfo `json:"workload"`
	HPA           HPAInfo      `json:"hpa"`
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
	var hpaUtils []float64

	for _, project := range list.Data {
		// Namespace
		parts := strings.SplitN(project.ID, ":", 2)
		if len(parts) == 2 {
			clusterID := parts[0]
			projectID := parts[1]
			cluster, err := c.Client.Cluster.ByID(clusterID)
			if err != nil {
				log.Errorf("Failed to get cluster %s err=%s", clusterID, err)
				return nil
			}
			nsCollection := filterNSCollectionWithProjectID(GetNamespaceCollection(c, cluster.Links["namespaces"]), projectID)
			totalNs := len(nsCollection.Data)
			p.Ns.Update(totalNs)
			nsUtils = append(nsUtils, float64(totalNs))
			p.Ns.UpdateDetails(&nsCollection)
		}
		// Workload
		wlCollection := GetWorkloadCollection(c, project.Links["workloads"])
		if wlCollection != nil {
			totalWl := len(wlCollection.Data)
			p.Workload.Update(totalWl)
			wlUtils = append(wlUtils, float64(totalWl))
		}

		// HPA
		hpaCollection := GetHPACollection(c, project.Links["horizontalPodAutoscalers"])
		if hpaCollection != nil {
			totalHPAs := len(hpaCollection.Data)
			p.HPA.Update(totalHPAs)
			hpaUtils = append(hpaUtils, float64(totalHPAs))
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
	p.HPA.UpdateAvg(hpaUtils)
	p.Pod.UpdateAvg(poUtils)

	return p
}

func filterNSCollectionWithProjectID(collection *rancher.NamespaceCollection, projectID string) rancher.NamespaceCollection {
	result := rancher.NamespaceCollection{
		Data: []rancher.Namespace{},
	}
	if collection == nil {
		return result
	}
	for _, ns := range collection.Data {
		if ns.Labels[projectLabel] == projectID {
			result.Data = append(result.Data, ns)
		}
	}
	return result
}

func init() {
	Register(Project{})
}

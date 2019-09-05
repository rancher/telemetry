package collector

import (
	"fmt"
	"net/url"
	"strings"

	rancher "github.com/rancher/types/client/cluster/v3"
	log "github.com/sirupsen/logrus"
)

const orchestrationName = "cattle-V2.0"
const rancherCatalogURL = "https://git.rancher.io/charts"
const projectLabel = "field.cattle.io/projectId"

type Project struct {
	Total         int          `json:"total"`
	Ns            NsInfo       `json:"namespace"`
	Workload      WorkloadInfo `json:"workload"`
	Pipeline      PipelineInfo `json:"pipeline"`
	LibraryCharts LabelCount   `json:"charts"`
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

	p.LibraryCharts = make(LabelCount)
	p.Orchestration = make(LabelCount)
	p.Orchestration[orchestrationName] = total
	p.Total = total

	var nsUtils []float64
	var wlUtils []float64
	var poUtils []float64
	var hpaUtils []float64

	// Setup vars for catalogs
	perClusterCatalogMap := make(map[string]bool)
	rancherCatalog, err := c.Client.Catalog.ByID("library")
	if err != nil || rancherCatalog.URL != rancherCatalogURL {
		log.Error("Failed to find a valid rancher default catalog")
		rancherCatalog = nil
	}

	for _, project := range list.Data {
		// Namespace
		parts := strings.SplitN(project.ID, ":", 2)
		if len(parts) == 2 {
			clusterID := parts[0]
			projectID := parts[1]
			cluster, err := c.Client.Cluster.ByID(clusterID)
			if err != nil {
				log.Errorf("Failed to get cluster %s for project %s err=%s", clusterID, projectID, err)
				continue
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

		// Pipeline
		pipelineCollection := GetPipelineCollection(c, project.Links["pipelines"])
		if pipelineCollection != nil {
			p.Pipeline.TotalPipelines += len(pipelineCollection.Data)
		}

		// Source provider
		p.Pipeline.SourceProvider = make(LabelCount)
		sourceCollection := GetSourceCodeProviderCollection(c, project.Links["sourceCodeProviders"])
		if sourceCollection != nil {
			p.Pipeline.Enabled = 1
			for _, provider := range sourceCollection.Data {
				p.Pipeline.SourceProvider.Increment(provider.Type)
			}
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

		// Apps
		if len(parts) == 2 {
			clusterID := parts[0]
			if rancherCatalog != nil {
				appsCollection := GetAppsCollection(c, project.Links["apps"])
				if appsCollection != nil {
					for _, app := range appsCollection.Data {
						catalog, catalogType, template, err := SplitExternalID(app.ExternalID)
						if err != nil {
							log.Debugf("Could not parse ExternalID %s", app.ExternalID)
							continue
						}
						if catalog == rancherCatalog.Name && catalogType != "clusterCatalog" && catalogType != "projectCatalog" {
							perClusterKey := fmt.Sprintf("%s:%s", clusterID, template) // Only count 1 per cluster
							if perClusterCatalogMap[perClusterKey] != true {
								perClusterCatalogMap[perClusterKey] = true
								p.LibraryCharts.Increment(template)
							}
						}
					}
				}
			}
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

func SplitExternalID(externalID string) (string, string, string, error) {
	var templateVersionNamespace, catalog string
	values, err := url.Parse(externalID)
	if err != nil {
		return "", "", "", err
	}
	catalogWithNamespace := values.Query().Get("catalog")
	catalogType := values.Query().Get("type")
	template := values.Query().Get("template")
	split := strings.SplitN(catalogWithNamespace, "/", 2)
	if len(split) == 2 {
		templateVersionNamespace = split[0]
		catalog = split[1]
	}
	// pre-upgrade setups will have global catalogs, where externalId field on templateversions won't have namespace.
	// since these are global catalogs, we can default to global namespace
	if templateVersionNamespace == "" {
		templateVersionNamespace = "cattle-global-data"
		catalog = catalogWithNamespace
	}
	return catalog, catalogType, template, nil
}

func init() {
	Register(Project{})
}

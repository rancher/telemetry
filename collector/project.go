package collector

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const orchestrationName = "cattle-V2.0"
const rancherCatalogURL = "https://git.rancher.io/charts"

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

	nonRemoved := NonRemoved()

	log.Debug("Collecting Projects")
	list, err := c.Client.Project.ListAll(&opts)

	if err != nil {
		log.Errorf("Failed to get Projects err=%s", err)
		return nil
	}

	total := len(list.Data)
	log.Debugf("  Found %d Projects", total)

	p.LibraryCharts = make(LabelCount)
	p.Orchestration = make(LabelCount)
	p.Pipeline.SourceProvider = make(LabelCount)
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
		parts := strings.SplitN(project.ID, ":", 2)
		clusterID := parts[0]
		clusterClient, err := GetClusterClient(c, clusterID)
		if err != nil {
			log.Errorf("Failed to get cluster client ID %s err=%s", clusterID, err)
		} else {
			// Namespace
			log.Debugf("  Collecting namespaces")
			nsFilter := NonRemoved()
			nsFilter.Filters["projectId"] = project.ID
			nsCollection, err := clusterClient.Namespace.ListAll(&nsFilter)
			if err != nil {
				log.Errorf("Failed to get Namespaces for project %s err=%s", project.ID, err)
			} else {
				totalNs := len(nsCollection.Data)
				p.Ns.Update(totalNs)
				nsUtils = append(nsUtils, float64(totalNs))
				p.Ns.UpdateDetails(nsCollection.Data)
				log.Debugf("    Found %d namespaces", totalNs)
			}
		}

		projectClient, err := GetProjectClient(c, project.ID)
		if err != nil {
			log.Errorf("Failed to get project client ID %s err=%s", project.ID, err)
			continue
		}

		// Workload
		log.Debugf("  Collecting Workloads")
		wlCollection, err := projectClient.Workload.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get Workload for project %s err=%s", project.ID, err)
		} else {
			totalWl := len(wlCollection.Data)
			p.Workload.Update(totalWl)
			wlUtils = append(wlUtils, float64(totalWl))
			log.Debugf("    Found %d Workloads", totalWl)
		}

		// Pipeline
		log.Debugf("  Collecting Pipelines")
		pipelineCollection, err := projectClient.Pipeline.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get Pipelines for project %s err=%s", project.ID, err)
		} else {
			p.Pipeline.TotalPipelines += len(pipelineCollection.Data)
			log.Debugf("    Found %d Pipelines", p.Pipeline.TotalPipelines)
		}

		// Source provider
		log.Debugf("  Collecting SourceCodeProviders")
		sourceCollection, err := projectClient.SourceCodeProvider.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get SourceCodeProvider for project %s err=%s", project.ID, err)
		} else {
			p.Pipeline.Enabled = 1
			for _, provider := range sourceCollection.Data {
				p.Pipeline.SourceProvider.Increment(provider.Type)
			}
			log.Debugf("    Found %d SourceCodeProviders", len(sourceCollection.Data))
		}

		// HPA
		log.Debugf("  Collecting HPAs")
		hpaCollection, err := projectClient.HorizontalPodAutoscaler.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get HPA for project %s err=%s", project.ID, err)
		} else {
			totalHPAs := len(hpaCollection.Data)
			p.HPA.Update(totalHPAs)
			hpaUtils = append(hpaUtils, float64(totalHPAs))
			log.Debugf("    Found %d HPAs", totalHPAs)
		}

		// Pod
		log.Debugf("  Collecting Pods")
		poCollection, err := projectClient.Pod.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get Pod for project %s err=%s", project.ID, err)
		} else {
			totalPo := len(poCollection.Data)
			p.Pod.Update(totalPo)
			poUtils = append(poUtils, float64(totalPo))
			log.Debugf("    Found %d Pods", totalPo)
		}

		// Apps
		if len(parts) == 2 {
			clusterID := parts[0]
			if rancherCatalog != nil {
				log.Debugf("  Collecting Apps")
				appsCollection, err := projectClient.App.ListAll(&nonRemoved)
				if err != nil {
					log.Errorf("Failed to get Apps for project %s err=%s", project.ID, err)
				} else {
					for _, app := range appsCollection.Data {
						catalog, catalogType, template, err := SplitExternalID(app.ExternalID)
						if err != nil {
							log.Debugf("Could not parse ExternalID %s", app.ExternalID)
							continue
						}
						if catalog == rancherCatalog.Name && catalogType != "clusterCatalog" && catalogType != "projectCatalog" {
							perClusterKey := fmt.Sprintf("%s:%s", clusterID, template) // Only count 1 per cluster
							if !perClusterCatalogMap[perClusterKey] {
								perClusterCatalogMap[perClusterKey] = true
								p.LibraryCharts.Increment(template)
							}
						}
					}
					log.Debugf("    Found %d Apps", len(appsCollection.Data))
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
		catalog = catalogWithNamespace
	}
	return catalog, catalogType, template, nil
}

func init() {
	Register(Project{})
}

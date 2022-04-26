package collector

import (
	"fmt"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	v1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	k3sEmbeddedDriver  = "k3sBased"
	k3sRancherDeploy   = "rancher"
	k3sRancherDeployNs = "cattle-system"
	systemProjectLabel = "authz.management.cattle.io/system-project"
)

type Cluster struct {
	Active             int          `json:"active"`
	Total              int          `json:"total"`
	Ns                 *NsInfo      `json:"namespace"`
	Project            *ProjectInfo `json:"project"`
	Cpu                *CpuInfo     `json:"cpu"`
	Mem                *MemoryInfo  `json:"mem"`
	Pod                *PodInfo     `json:"pod"`
	Driver             LabelCount   `json:"driver"`
	SecretTotal        int          `json:"secret"`
	IstioTotal         int          `json:"istio"`
	MonitoringTotal    int          `json:"monitoring"`
	Service            *ServiceInfo `json:"service"`
	LogProviderCount   LabelCount   `json:"logging"`
	CloudProvider      LabelCount   `json:"cloudProvider"`
	CISTotal           int          `json:"cis"`
	CISProfileVersion  LabelCount   `json:"cisProfileVersion"`
	NetworkPolicyTotal int          `json:"networkPolicy"`
	Hybrid             int          `json:"hybrid"`
	LinuxWorkerTotal   int          `json:"linuxWorker"`
	WindowsWorkerTotal int          `json:"windowsWorker"`
}

func (h Cluster) RecordKey() string {
	return "cluster"
}

func (h Cluster) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Clusters")
	clusterList, err := c.Client.Cluster.ListAll(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Clusters err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Clusters", len(clusterList.Data))

	h.Ns = &NsInfo{}
	h.Cpu = &CpuInfo{}
	h.Mem = &MemoryInfo{}
	h.Pod = &PodInfo{}
	h.Driver = make(LabelCount)
	h.CloudProvider = make(LabelCount)

	var cpuUtils []float64
	var memUtils []float64
	var podUtils []float64
	var nsUtils []float64
	var serviceUtils []float64

	// Clusters
	for _, cluster := range clusterList.Data {
		var utilFloat float64
		var util int

		log.Debugf("  Cluster: %s", displayClusterName(cluster))

		h.Total++
		if cluster.State == "active" {
			h.Active++
		}

		allocatable := cluster.Allocatable
		totalCores := GetCPU(allocatable["cpu"])
		totalMemMb := GetMemMb(allocatable["memory"])
		totalPods := GetRawInt(allocatable["pods"], "")
		if totalCores == 0 || totalMemMb == 0 || totalPods == 0 {
			log.Debugf("  Skipping Cluster with no resources: %s", displayClusterName(cluster))
			continue
		}

		requested := cluster.Requested

		// CPU
		usedCores := GetRawInt(requested["cpu"], "m")
		utilFloat = float64(usedCores) / float64(totalCores*10)
		util = Round(utilFloat)

		h.Cpu.Update(totalCores, util)
		cpuUtils = append(cpuUtils, utilFloat)
		log.Debugf("    CPU cores=%d, util=%d", totalCores, util)

		// Memory
		usedMemMB := GetMemMb(requested["memory"])
		utilFloat = 100 * float64(usedMemMB) / float64(totalMemMb)
		util = Round(utilFloat)

		h.Mem.Update(totalMemMb, util)
		memUtils = append(memUtils, utilFloat)
		log.Debugf("    Mem used=%d, total=%d, util=%d", usedMemMB, totalMemMb, util)

		// Pod
		usedPods := GetRawInt(requested["pods"], "")
		utilFloat = 100 * float64(usedPods) / float64(totalPods)
		util = Round(utilFloat)

		h.Pod.Update(totalPods, util)
		podUtils = append(podUtils, utilFloat)
		log.Debugf("    Pod used=%d, total=%d, util=%d", usedPods, totalPods, util)

		// Driver
		// Check if Rancher is running on embedded k3s
		if isK3sEmbedded(c, cluster) {
			h.Driver.Increment(k3sEmbeddedDriver)
		} else {
			h.Driver.Increment(cluster.Driver)
		}

		// Windows Worker Node Count
		if cluster.WindowsWorkerCount > 0 {
			h.Hybrid++
			utilFloat = float64(cluster.WindowsWorkerCount)
			util = Round(utilFloat)
			h.WindowsWorkerTotal += util

		}

		// Linux Worker Node Count
		if cluster.LinuxWorkerCount > 0 {
			utilFloat = float64(cluster.LinuxWorkerCount)
			util = Round(utilFloat)
			h.LinuxWorkerTotal += util
		}

		// CIS
		if cluster.ScheduledClusterScan.Enabled {
			h.CISTotal++
			h.CISProfileVersion.Increment(cluster.ScheduledClusterScan.ScanConfig.CisScanConfig.Profile)
		}

		// Network Policy
		if cluster.AppliedEnableNetworkPolicy {
			h.NetworkPolicyTotal++
		}

		if cluster.RancherKubernetesEngineConfig != nil && cluster.RancherKubernetesEngineConfig.CloudProvider != nil {
			if cluster.RancherKubernetesEngineConfig.CloudProvider.Name != "" {
				h.CloudProvider.Increment(
					cluster.RancherKubernetesEngineConfig.CloudProvider.Name)
			}
		}

		var secretClient v1.SecretClient
		secretList, err := secretClient.List("", metav1.ListOptions{})
		if err == nil {
			for range secretList.Items {
				h.SecretTotal++
			}
		}

		// Namespace
		clusterClient, err := GetClusterClient(c, cluster.ID)
		if err != nil {
			log.Errorf("Failed to get Cluster client err=%s", err)
		}
		nsCollection, err := clusterClient.Namespace.ListAll(nil)
		if err != nil {
			log.Errorf("Failed to get Namespaces err=%s", err)
		} else {
			totalNs := len(nsCollection.Data)
			h.Ns.Update(totalNs)
			nsUtils = append(nsUtils, float64(totalNs))
			h.Ns.UpdateDetails(nsCollection.Data)
		}

		// Secret
		//secretCollection := v1.SecretLister()
		//if err != nil {
		//	log.Errorf("Failed to get Secrets err=%s", err)
		//} else {
		//	totalSecret := len(secretCollection.Data)
		//	h.SecretTotal.Update(totalSecret)
		//	nsUtils = append(nsUtils, float64(totalSecret))
		//	h.Ns.UpdateDetails(secretCollection.Data)
		//}

		// Services
		serviceCollection, err := clusterClient.APIService.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get Services err=%s", err)
		} else {
			totalServices := len(serviceCollection.Data)
			h.Service.Update(totalServices)
			serviceUtils = append(serviceUtils, float64(totalServices))
			h.Service.UpdateDetails(serviceCollection.Data)
		}

		// Monitoring
		if cluster.EnableClusterMonitoring {
			h.MonitoringTotal++
		}

		// Istio
		if cluster.IstioEnabled {
			h.IstioTotal++
		}
	}

	h.Cpu.UpdateAvg(cpuUtils)
	h.Mem.UpdateAvg(memUtils)
	h.Pod.UpdateAvg(podUtils)
	h.Ns.UpdateAvg(nsUtils)

	// Cluster Logging
	h.LogProviderCount = make(LabelCount)

	logList, err := c.Client.ClusterLogging.ListAll(nil)
	if err != nil {
		log.Errorf("Failed to get Cluster Loggings err=%s", err)
		return nil
	}

	for _, logging := range logList.Data {
		if logging.AppliedSpec != nil {
			switch {
			case logging.AppliedSpec.ElasticsearchConfig != nil:
				h.LogProviderCount["Elasticsearch"]++
			case logging.AppliedSpec.SplunkConfig != nil:
				h.LogProviderCount["Splunk"]++
			case logging.AppliedSpec.KafkaConfig != nil:
				h.LogProviderCount["Kafka"]++
			case logging.AppliedSpec.SyslogConfig != nil:
				h.LogProviderCount["Syslog"]++
			case logging.AppliedSpec.FluentForwarderConfig != nil:
				h.LogProviderCount["Fluentd"]++
			case logging.AppliedSpec.CustomTargetConfig != nil:
				h.LogProviderCount["Custom"]++
			}
		}
	}

	return h
}

func init() {
	Register(Cluster{})
}

func displayClusterName(c rancher.Cluster) string {
	if len(c.Name) > 0 {
		return c.Name
	} else {
		return "(" + c.UUID + ")"
	}
}

func isK3sEmbedded(c *CollectorOpts, cluster rancher.Cluster) bool {
	if cluster.Driver == "k3s" && cluster.Internal {
		systemProjectID, err := getClusterSystemProjectID(c, cluster.ID)
		if err != nil {
			log.Debugf("Failed to get System project ID err=%s", err)
			return false
		}

		// Checking if Rancher is running as workload within the cluster
		projectCli, err := GetProjectClient(c, systemProjectID)
		listOpts := NonRemoved()
		listOpts.Filters["name"] = k3sRancherDeploy
		listOpts.Filters["namespaceId"] = k3sRancherDeployNs
		projects, err := projectCli.Workload.List(&listOpts)
		if err != nil {
			log.Debugf("Failed to get System project deployments err=%s", err)
			return false
		}
		if len(projects.Data) == 0 {
			return true
		}
	}
	return false
}

func getClusterProjects(c *CollectorOpts, id string) ([]rancher.Project, error) {
	if id == "" {
		return nil, fmt.Errorf("[ERROR] Cluster id is nil")
	}

	listOpts := NonRemoved()
	listOpts.Filters["clusterId"] = id

	collection, err := c.Client.Project.List(&listOpts)
	if err != nil {
		return nil, err
	}

	return collection.Data, nil
}

func getClusterSystemProjectID(c *CollectorOpts, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("[ERROR] Cluster id is nil")
	}

	projects, err := getClusterProjects(c, id)
	if err != nil {
		return "", err
	}

	for _, project := range projects {
		if isSystemProject(&project) {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("[ERROR] System project not found at Cluster id %s", id)
}

func isSystemProject(project *rancher.Project) bool {
	if project == nil {
		return false
	}

	for k, v := range project.Labels {
		if k == systemProjectLabel && v == "true" {
			return true
		}
	}

	return false
}

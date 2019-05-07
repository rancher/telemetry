package collector

import (
	rancher "github.com/rancher/types/client/management/v3"
	log "github.com/sirupsen/logrus"
)

type Cluster struct {
	Active int         `json:"active"`
	Total  int         `json:"total"`
	Ns     *NsInfo     `json:"namespace"`
	Cpu    *CpuInfo    `json:"cpu"`
	Mem    *MemoryInfo `json:"mem"`
	Pod    *PodInfo    `json:"pod"`
	Driver LabelCount  `json:"driver"`
}

func (h Cluster) RecordKey() string {
	return "cluster"
}

func (h Cluster) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Clusters")
	clusterList, err := c.Client.Cluster.List(&nonRemoved)
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

	var cpuUtils []float64
	var memUtils []float64
	var podUtils []float64
	var nsUtils []float64

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
		totalCores := GetRawInt(allocatable["cpu"], "")
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
		h.Driver.Increment(cluster.Driver)

		// Namespace
		nsCollection := GetNamespaceCollection(c, cluster.Links["namespaces"])
		if nsCollection != nil {
			totalNs := len(nsCollection.Data)
			h.Ns.Update(totalNs)
			nsUtils = append(nsUtils, float64(totalNs))
			h.Ns.UpdateDetails(nsCollection)
		}
	}

	h.Cpu.UpdateAvg(cpuUtils)
	h.Mem.UpdateAvg(memUtils)
	h.Pod.UpdateAvg(podUtils)
	h.Ns.UpdateAvg(nsUtils)

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

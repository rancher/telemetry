package collector

import (
	rancher "github.com/rancher/types/client/management/v3"
	log "github.com/sirupsen/logrus"
)

type Node struct {
	Active    int        `json:"active"`
	Imported  int        `json:"imported"`
	FromTmpl  int        `json:"from_template"`
	Total     int        `json:"total"`
	Cpu       CpuInfo    `json:"cpu"`
	Mem       MemoryInfo `json:"mem"`
	Pod       PodInfo    `json:"pod"`
	Kernel    LabelCount `json:"kernel"`
	Kubelet   LabelCount `json:"kubelet"`
	Kubeproxy LabelCount `json:"kubeproxy"`
	Os        LabelCount `json:"os"`
	Docker    LabelCount `json:"docker"`
	Driver    LabelCount `json:"driver"`
	Role      LabelCount `json:"role"`
}

func (m Node) RecordKey() string {
	return "node"
}

func (h Node) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Nodes")
	nodeList, err := c.Client.Node.ListAll(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Nodes err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Nodes", len(nodeList.Data))

	var cpuUtils []float64
	var memUtils []float64
	var podUtils []float64

	h.Kernel = make(LabelCount)
	h.Kubelet = make(LabelCount)
	h.Kubeproxy = make(LabelCount)
	h.Os = make(LabelCount)
	h.Docker = make(LabelCount)
	h.Driver = make(LabelCount)
	h.Role = make(LabelCount)

	// Nodes
	for _, node := range nodeList.Data {
		var utilFloat float64
		var util int

		log.Debugf("  Node: %s", displayNodeName(node))

		h.Total++
		if node.State == "active" {
			h.Active++
		}
		if node.Imported {
			h.Imported++
		}

		allocatable := node.Allocatable
		totalCores := GetCPU(allocatable["cpu"])
		totalMemMb := GetMemMb(allocatable["memory"])
		totalPods := GetRawInt(allocatable["pods"], "")
		if totalCores == 0 || totalMemMb == 0 || totalPods == 0 {
			log.Debugf("  Skipping Node with no resources: %s", displayNodeName(node))
			continue
		}

		requested := node.Requested

		// CPU
		usedCores := GetRawInt(requested["cpu"], "m")
		if usedCores > 0 && totalCores > 0 {
			utilFloat = float64(usedCores) / float64(totalCores*10)
		}
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

		// OS
		osInfo := node.Info.OS
		h.Kernel.Increment(osInfo.KernelVersion)
		h.Os.Increment(osInfo.OperatingSystem)
		h.Docker.Increment(osInfo.DockerVersion)

		// K8s
		kubeInfo := node.Info.Kubernetes
		h.Kubelet.Increment(kubeInfo.KubeletVersion)
		h.Kubeproxy.Increment(kubeInfo.KubeProxyVersion)

		// Role
		if node.ControlPlane {
			h.Role.Increment("controlplane")
		}
		if node.Etcd {
			h.Role.Increment("etcd")
		}
		if node.Worker {
			h.Role.Increment("worker")
		}

		// Driver
		if len(node.NodeTemplateID) > 0 {
			nodeTemplate, err := c.Client.NodeTemplate.ByID(node.NodeTemplateID)
			if err != nil {
				if IsNotFound(err) {
					log.Debugf("    nodeTemplate not found [%s]", node.NodeTemplateID)
				} else {
					log.Errorf("Failed to get nodeTemplate [%s] err=%s", node.NodeTemplateID, err)
				}
			} else {
				h.FromTmpl++
				h.Driver.Increment(nodeTemplate.Driver)
			}
		}
	}

	h.Cpu.UpdateAvg(cpuUtils)
	h.Mem.UpdateAvg(memUtils)
	h.Pod.UpdateAvg(podUtils)

	return h
}

func init() {
	Register(Node{})
}

func displayNodeName(m rancher.Node) string {
	if len(m.Name) > 0 {
		return m.Name
	} else if len(m.Hostname) > 0 {
		return m.Hostname
	} else {
		return "(" + m.UUID + ")"
	}
}

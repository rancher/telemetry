package collector

import (
	rancher "github.com/rancher/types/client/management/v3"
	log "github.com/sirupsen/logrus"
)

type Machine struct {
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

func (m Machine) RecordKey() string {
	return "machine"
}

func (h Machine) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Machines")
	machineList, err := c.Client.Machine.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Machines err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Machines", len(machineList.Data))

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

	// Machines
	for _, machine := range machineList.Data {
		var utilFloat float64
		var util int

		log.Debugf("  Machine: %s", displayMachineName(machine))

		h.Total++
		if machine.State == "active" {
			h.Active++
		}
		if *machine.Imported {
			h.Imported++
		}

		allocatable := machine.Allocatable
		if allocatable["cpu"] == "0" || allocatable["memory"] == "0" || allocatable["pods"] == "0" {
			log.Debugf("  Skipping Machine with no resources: %s", displayMachineName(machine))
			continue
		}

		requested := machine.Requested

		// CPU
		totalCores := GetRawInt(allocatable["cpu"], "")
		usedCores := GetRawInt(requested["cpu"], "m")
		utilFloat = float64(usedCores) / float64(totalCores*10)
		util = Round(utilFloat)

		h.Cpu.Update(totalCores, util)
		cpuUtils = append(cpuUtils, utilFloat)
		log.Debugf("    CPU cores=%d, util=%d", totalCores, util)

		// Memory
		totalMemMb := GetMemMb(allocatable["memory"])
		usedMemMB := GetMemMb(requested["memory"])
		utilFloat = 100 * float64(usedMemMB) / float64(totalMemMb)
		util = Round(utilFloat)

		h.Mem.Update(totalMemMb, util)
		memUtils = append(memUtils, utilFloat)
		log.Debugf("    Mem used=%d, total=%d, util=%d", usedMemMB, totalMemMb, util)

		// Pod
		totalPods := GetRawInt(allocatable["pods"], "")
		usedPods := GetRawInt(requested["pods"], "")
		utilFloat = 100 * float64(usedPods) / float64(totalPods)
		util = Round(utilFloat)

		h.Pod.Update(totalPods, util)
		podUtils = append(podUtils, utilFloat)
		log.Debugf("    Pod used=%d, total=%d, util=%d", usedPods, totalPods, util)

		// OS
		osInfo := machine.Info.OS
		h.Kernel.Increment(osInfo.KernelVersion)
		h.Os.Increment(osInfo.OperatingSystem)
		h.Docker.Increment(osInfo.DockerVersion)

		// K8s
		kubeInfo := machine.Info.Kubernetes
		h.Kubelet.Increment(kubeInfo.KubeletVersion)
		h.Kubeproxy.Increment(kubeInfo.KubeProxyVersion)

		// Role
		for _, role := range machine.Role {
			h.Role.Increment(role)
		}

		// Driver
		machineTemplate := GetMachineTemplate(c.Client, machine.MachineTemplateId)
		if machineTemplate != nil {
			h.FromTmpl++
			h.Driver.Increment(machineTemplate.Driver)
		}
	}

	h.Cpu.UpdateAvg(cpuUtils)
	h.Mem.UpdateAvg(memUtils)
	h.Pod.UpdateAvg(podUtils)

	return h
}

func init() {
	Register(Machine{})
}

func displayMachineName(m rancher.Machine) string {
	if len(m.Name) > 0 {
		return m.Name
	} else if len(m.Hostname) > 0 {
		return m.Hostname
	} else {
		return "(" + m.Uuid + ")"
	}
}

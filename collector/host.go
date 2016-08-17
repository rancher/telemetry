package collector

import (
	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/client"
	"regexp"
)

type CpuInfo struct {
	CoresMin   int `json:"cores_min"`
	CoresMax   int `json:"cores_max"`
	CoresTotal int `json:"cores_total"`
	MhzTotal   int `json:"mhz_total"`
	UtilMin    int `json:"util_min"`
	UtilAvg    int `json:"util_avg"`
	UtilMax    int `json:"util_max"`
}

type MemoryInfo struct {
	MinMb   int `json:"mb_min"`
	MaxMb   int `json:"mb_max"`
	TotalMb int `json:"mb_total"`
	UtilMin int `json:"util_min"`
	UtilAvg int `json:"util_avg"`
	UtilMax int `json:"util_max"`
}

type Host struct {
	Count      int        `json:"count"`
	Cpu        CpuInfo    `json:"cpu"`
	Mem        MemoryInfo `json:"mem"`
	Kernel     LabelCount `json:"kernel"`
	Os         LabelCount `json:"os"`
	DockerFull LabelCount `json:"docker_full"`
	Docker     LabelCount `json:"docker"`
	Driver     LabelCount `json:"driver"`
}

func (h Host) RecordKey() string {
	return "host"
}

func (h Host) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Hosts")
	hostList, err := c.Client.Host.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Hosts err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Hosts", len(hostList.Data))

	var cpuUtils []float64
	var memUtils []float64

	h.Kernel = make(LabelCount)
	h.Os = make(LabelCount)
	h.Docker = make(LabelCount)
	h.DockerFull = make(LabelCount)
	h.Driver = make(LabelCount)

	// Hosts
	for _, host := range hostList.Data {
		var utilFloat float64
		var util int

		info := host.Info.(map[string]interface{})
		log.Debugf("  Host: %s", displayName(host))

		// CPU
		cpuInfo := info["cpuInfo"].(map[string]interface{})
		cores := Round(cpuInfo["count"].(float64))
		mhz := Round(cpuInfo["mhz"].(float64))

		for _, core := range cpuInfo["cpuCoresPercentages"].([]interface{}) {
			utilFloat += core.(float64)
		}
		utilFloat = utilFloat / float64(cores)
		util = Round(utilFloat)

		h.Count += 1
		h.Cpu.CoresMin = MinButNotZero(h.Cpu.CoresMin, cores)
		h.Cpu.CoresMax = Max(h.Cpu.CoresMin, cores)
		h.Cpu.CoresTotal += cores
		h.Cpu.MhzTotal += mhz

		h.Cpu.UtilMin = MinButNotZero(h.Cpu.UtilMin, util)
		h.Cpu.UtilMax = Max(h.Cpu.UtilMax, util)
		cpuUtils = append(cpuUtils, utilFloat)
		log.Debugf("    CPU cores=%d, mhz=%d, util=%d", cores, mhz, util)

		// Memory
		memInfo := info["memoryInfo"].(map[string]interface{})
		total := Round(memInfo["memTotal"].(float64))
		avail := Round(memInfo["memAvailable"].(float64))
		used := Clamp(0, total-avail, total)
		utilFloat = 100 * float64(used) / float64(total)
		util = Round(utilFloat)

		h.Mem.MinMb = MinButNotZero(h.Mem.MinMb, total)
		h.Mem.MaxMb = Max(h.Mem.MaxMb, total)
		h.Mem.TotalMb += total
		h.Mem.UtilMin = MinButNotZero(h.Mem.UtilMin, util)
		h.Mem.UtilMax = Max(h.Mem.UtilMax, util)
		memUtils = append(memUtils, utilFloat)
		log.Debugf("    Mem used=%d, total=%d, util=%d", used, total, util)

		// OS
		osInfo := info["osInfo"].(map[string]interface{})
		h.Kernel.Increment(osInfo["kernelVersion"].(string))
		h.Os.Increment(osInfo["operatingSystem"].(string))

		dockerFull := osInfo["dockerVersion"].(string)
		docker := regexp.MustCompile("(?i)^docker version (.*)").ReplaceAllString(dockerFull, "v$1")
		docker = regexp.MustCompile("(?i)^(.*), build [0-9a-f]+$").ReplaceAllString(docker, "$1")
		h.DockerFull.Increment(dockerFull)
		h.Docker.Increment(docker)
	}

	h.Cpu.UtilAvg = Clamp(0, Round(Average(cpuUtils)), 100)
	h.Mem.UtilAvg = Clamp(0, Round(Average(memUtils)), 100)

	// Machine Drivers
	log.Debug("  Collecting Machines")
	machineList, err := c.Client.Machine.List(&nonRemoved)

	if err != nil {
		log.Errorf("Failed to get Machines err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Machines", len(machineList.Data))
	for _, machine := range machineList.Data {
		h.Driver.Increment(machine.Driver)
	}
	h.Driver["custom"] = Max(0, h.Count-len(machineList.Data))

	return h
}

func init() {
	Register(Host{})
}

func displayName(h rancher.Host) string {
	if len(h.Name) > 0 {
		return h.Name
	} else if len(h.Hostname) > 0 {
		return h.Hostname
	} else {
		return "(" + h.Id + ")"
	}
}

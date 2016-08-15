package collector

import (
	log "github.com/Sirupsen/logrus"
)

type Container struct {
	Running    int `json:"running"`
	Total      int `json:"total"`
	PerHostMin int `json:"per_host_min"`
	PerHostAvg int `json:"per_host_avg"`
	PerHostMax int `json:"per_host_max"`
}

func (c Container) RecordKey() string {
	return "container"
}

func (out Container) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Container")
	list, err := c.Client.Container.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Containers err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Containers", len(list.Data))

	byHost := make(LabelCount)

	total := len(list.Data)
	out.Total = total

	for _, container := range list.Data {
		if container.State == "running" {
			out.Running++
		}
		byHost.Increment(container.HostId)
	}

	var flat []float64
	for hostId, count := range byHost {
		log.Debugf("  Host id=%s, count=%d", hostId, count)
		out.PerHostMin = MinButNotZero(out.PerHostMin, count)
		out.PerHostMax = Max(out.PerHostMax, count)
		flat = append(flat, float64(count))
	}

	out.PerHostAvg = Round(Average(flat))

	return out
}

func init() {
	Register(Container{})
}

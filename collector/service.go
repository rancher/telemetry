package collector

import (
	log "github.com/Sirupsen/logrus"
)

type Service struct {
	Active      int            `json:"active"`
	Total       int            `json:"total"`
	Kind        map[string]int `json:"kind"`
	PerStackMin int            `json:"per_stack_min"`
	PerStackAvg int            `json:"per_stack_avg"`
	PerStackMax int            `json:"per_stack_max"`
}

func (s Service) RecordKey() string {
	return "service"
}

func (out Service) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Service")
	list, err := c.Client.Service.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Services err=%s", err)
		return nil
	}

	log.Debugf("  Found %d Services", len(list.Data))

	byStack := make(map[string]int)
	total := len(list.Data)

	out.Kind = make(map[string]int)
	out.Total = total

	for _, service := range list.Data {
		if service.State == "active" {
			out.Active++
		}

		IncrementMap(&byStack, service.EnvironmentId)
		IncrementMap(&out.Kind, service.Type)
	}

	var flat []float64
	for stackId, count := range byStack {
		log.Debugf("  Stack id=%s, count=%d", stackId, count)
		out.PerStackMin = MinButNotZero(out.PerStackMin, count)
		out.PerStackMax = Max(out.PerStackMax, count)
		flat = append(flat, float64(count))
	}

	out.PerStackAvg = Round(Average(flat))

	return out
}

func init() {
	Register(Service{})
}

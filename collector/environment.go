package collector

import (
	log "github.com/Sirupsen/logrus"
)

type Environment struct {
	Total         int        `json:"total"`
	Orchestration LabelCount `json:"orch"`
}

func (s Environment) RecordKey() string {
	return "environment"
}

func (out Environment) Collect(c *CollectorOpts) interface{} {
	opts := NonRemoved()
	opts.Filters["all"] = "true"

	log.Debug("Collecting Environment")
	list, err := c.Client.Project.List(&opts)

	if err != nil {
		log.Errorf("Failed to get Environments err=%s", err)
		return nil
	}

	total := len(list.Data)
	log.Debugf("  Found %d Environments", total)

	out.Orchestration = make(LabelCount)
	out.Total = total

	for _, env := range list.Data {
		// Enviornments can technically have more than one of these set...
		orch := env.Orchestration
		if len(orch) == 0 {
			orch = "cattle"
		}

		out.Orchestration.Increment(orch)
	}

	return out
}

func init() {
	Register(Environment{})
}

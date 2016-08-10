package collector

import (
	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/client"
)

type Environment struct {
	Total         int            `json:"total"`
	Orchestration map[string]int `json:"orch"`
}

func (s Environment) RecordKey() string {
	return "environment"
}

func (out Environment) Collect(c *CollectorOpts) interface{} {
	filters := make(map[string]interface{})
	filters["all"] = "true"

	log.Debug("Collecting Environment")
	list, err := c.Client.Project.List(&rancher.ListOpts{
		Filters: filters,
	})

	if err != nil {
		log.Errorf("Failed to get Environments err=%s", err)
		return nil
	}

	total := len(list.Data)
	log.Debugf("  Found %d Environments", total)

	out.Orchestration = make(map[string]int)
	out.Total = total

	for _, env := range list.Data {
		// Enviornments can technically have more than one of these set...
		found := false
		if env.Kubernetes {
			IncrementMap(&out.Orchestration, "kubernetes")
			found = true
		} else if env.Swarm {
			IncrementMap(&out.Orchestration, "swarm")
			found = true
		} else if env.Mesos {
			IncrementMap(&out.Orchestration, "mesos")
			found = true
		}

		if !found {
			IncrementMap(&out.Orchestration, "cattle")
		}
	}

	return out
}

func init() {
	Register(Environment{})
}

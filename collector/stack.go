package collector

import (
	log "github.com/Sirupsen/logrus"
	"strings"
)

type Stack struct {
	Active      int `json:"active"`
	FromCatalog int `json:"from_catalog"`
	Total       int `json:"total"`
	PerEnvMin   int `json:"per_env_min"`
	PerEnvAvg   int `json:"per_env_avg"`
	PerEnvMax   int `json:"per_env_max"`
}

func (s Stack) RecordKey() string {
	return "stack"
}

func (out Stack) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	log.Debug("Collecting Stack")
	list, err := c.Client.Stack.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Stacks err=%s", err)
		return nil
	}

	total := len(list.Data)
	log.Debugf("  Found %d Stacks", total)

	byEnvironment := make(LabelCount)

	out.Total = total

	for _, stack := range list.Data {
		if stack.State == "active" {
			out.Active++
		}

		if strings.Contains(stack.ExternalId, "catalog://") {
			out.FromCatalog++
		}

		byEnvironment.Increment(stack.AccountId)
	}

	var flat []float64
	for envId, count := range byEnvironment {
		log.Debugf("  Environment id=%s, count=%d", envId, count)
		out.PerEnvMin = MinButNotZero(out.PerEnvMin, count)
		out.PerEnvMax = Max(out.PerEnvMax, count)
		flat = append(flat, float64(count))
	}

	out.PerEnvAvg = Round(Average(flat))

	return out
}

func init() {
	Register(Stack{})
}

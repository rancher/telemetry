package collector

import (
	norman "github.com/rancher/norman/types"
	rancher "github.com/rancher/types/client/project/v3"
	log "github.com/sirupsen/logrus"
)

type HPAInfo struct {
	Min   int `json:"min"`
	Max   int `json:"max"`
	Total int `json:"total"`
	Avg   int `json:"avg"`
}

func (h *HPAInfo) Update(i int) {
	h.Total += i
	h.Min = MinButNotZero(h.Min, i)
	h.Max = Max(h.Max, i)
}

func (h *HPAInfo) UpdateAvg(i []float64) {
	h.Avg = Clamp(0, Round(Average(i)), 100)
}

func GetHPACollection(c *CollectorOpts, url string) *rancher.HorizontalPodAutoscalerCollection {
	if url == "" {
		log.Debugf("HPA collection link is empty.")
		return nil
	}

	hpaCollection := &rancher.HorizontalPodAutoscalerCollection{}
	version := "horizontalpodautoscalers"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, hpaCollection)
	if err != nil {
		log.Errorf("Error getting hpa collection [%s] %s", resource.Links[version], err)
		return nil
	}
	if hpaCollection == nil || hpaCollection.Type != "collection" || len(hpaCollection.Data) == 0 {
		log.Debugf("HPA collection is empty [%s]", resource.Links[version])
		return nil
	}

	return hpaCollection
}

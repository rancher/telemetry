package collector

import (
	norman "github.com/rancher/norman/types"
	rancher "github.com/rancher/types/client/project/v3"
	log "github.com/sirupsen/logrus"
)

type PipelineInfo struct {
	Enabled        int        `json:"enabled"` // 1 if user has any # pipeline provider enabled
	SourceProvider LabelCount `json:"source"`
	TotalPipelines int        `json:"total"`
}

func GetPipelineCollection(c *CollectorOpts, url string) *rancher.PipelineCollection {
	if url == "" {
		log.Debugf("Pipeline collection link is empty.")
		return nil
	}

	pipelineCollection := &rancher.PipelineCollection{}
	version := "pipeline"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, pipelineCollection)
	if err != nil {
		log.Errorf("Error getting pipeline collection [%s] %s", resource.Links[version], err)
		return nil
	}

	if pipelineCollection == nil || pipelineCollection.Type != "collection" || len(pipelineCollection.Data) == 0 {
		log.Debugf("Pipeline collection is empty [%s]", resource.Links[version])
		return nil
	}

	return pipelineCollection
}

func GetSourceCodeProviderCollection(c *CollectorOpts, url string) *rancher.SourceCodeProviderCollection {
	if url == "" {
		log.Debugf("SourceCodeProvier collection link is empty.")
		return nil
	}

	providerCollection := &rancher.SourceCodeProviderCollection{}
	version := "sourceCodeProvider"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, providerCollection)
	if err != nil {
		log.Errorf("Error getting sourceCodeProvider collection [%s] %s", resource.Links[version], err)
		return nil
	}

	if providerCollection == nil || providerCollection.Type != "collection" || len(providerCollection.Data) == 0 {
		log.Debugf("SourceCodeProvider collection is empty [%s]", resource.Links[version])
		return nil
	}

	return providerCollection
}

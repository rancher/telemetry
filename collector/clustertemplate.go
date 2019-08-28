package collector

import (
	log "github.com/sirupsen/logrus"
)

type ClusterTemplate struct {
	TotalClusterTemplates  int    `json:"total"`
	TotalTemplateRevisions int    `json:"revisions"`
	Enforcement            string `json:"enforcement"`
}

func (ct ClusterTemplate) RecordKey() string {
	return "clustertemplate"
}

func (ct ClusterTemplate) Collect(c *CollectorOpts) interface{} {
	nonRemoved := NonRemoved()

	clusterTemplateList, err := c.Client.ClusterTemplate.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Clusters Templates err=%s", err)
		return nil
	}
	ct.TotalClusterTemplates = len(clusterTemplateList.Data)

	revisionsList, err := c.Client.ClusterTemplateRevision.List(&nonRemoved)
	if err != nil {
		log.Errorf("Failed to get Cluster Revisions err=%s", err)
		return nil
	}
	ct.TotalTemplateRevisions = len(revisionsList.Data)

	setting, err := c.Client.Setting.ByID("cluster-template-enforcement")
	if err != nil {
		log.Errorf("Failed to get setting in Clusters Templates collect err=%s", err)
		return nil
	}

	ct.Enforcement = setting.Default
	if setting.Value != "" {
		ct.Enforcement = setting.Value
	}

	return ct
}

func init() {
	Register(ClusterTemplate{})
}

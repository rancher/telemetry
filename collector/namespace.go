package collector

import (
	norman "github.com/rancher/norman/types"
	rancher "github.com/rancher/types/client/cluster/v3"
	log "github.com/sirupsen/logrus"
)

type NsInfo struct {
	NsMin       int `json:"min"`
	NsMax       int `json:"max"`
	NsTotal     int `json:"total"`
	NsAvg       int `json:"avg"`
	FromCatalog int `json:"from_catalog,omitempty"`
	NoProject   int `json:"no_project,omitempty"`
}

func (n *NsInfo) Update(i int) {
	n.NsTotal += i
	n.NsMin = MinButNotZero(n.NsMin, i)
	n.NsMax = Max(n.NsMax, i)
}

func (n *NsInfo) UpdateAvg(i []float64) {
	n.NsAvg = Clamp(0, Round(Average(i)), 100)
}

func (n *NsInfo) UpdateDetails(nsc *rancher.NamespaceCollection) {
	for _, ns := range nsc.Data {
		// ExternalID field is not on namespace definition yet
		//if FromCatalog(ns.ExternalID) {
		//	n.FromCatalog++
		//}
		if ns.ProjectID == "" {
			n.NoProject++
		}
	}
}

func GetNamespaceCollection(c *CollectorOpts, url string) *rancher.NamespaceCollection {
	if url == "" {
		log.Debugf("Namespace collection link is empty.")
		return nil
	}

	nsCollection := &rancher.NamespaceCollection{}
	version := "namespaces"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, nsCollection)
	if err != nil {
		log.Debugf("Error getting namespace collection [%s] %s", resource.Links[version], err)
		return nil
	}
	if nsCollection == nil || nsCollection.Type != "collection" || len(nsCollection.Data) == 0 {
		log.Debugf("Namespace collection is empty [%s]", resource.Links[version])
		return nil
	}

	return nsCollection
}

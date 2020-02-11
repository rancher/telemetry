package collector

import (
	rancher "github.com/rancher/types/client/cluster/v3"
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

func (n *NsInfo) UpdateDetails(nsc []rancher.Namespace) {
	for _, ns := range nsc {
		// ExternalID field is not on namespace definition yet
		//if FromCatalog(ns.ExternalID) {
		//	n.FromCatalog++
		//}
		if ns.ProjectID == "" {
			n.NoProject++
		}
	}
}

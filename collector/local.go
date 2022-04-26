package collector

import (
	_ "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
)

type LocalClusterInfo struct {
	NsTotal      int `json:"local_ns_total"`
	ProjectTotal int `json:"local_project_total,omitempty"`
}

func (l *LocalClusterInfo) Update(i int) {
	l.NsTotal += i
	l.ProjectTotal += i
}

package collector

import (
	rancher "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
)

type ServiceInfo struct {
	ServiceMin   int `json:"min"`
	ServiceMax   int `json:"max"`
	ServiceTotal int `json:"total"`
	ServiceAvg   int `json:"avg"`
	NoProject    int `json:"no_project,omitempty"`
}

func (s *ServiceInfo) Update(i int) {
	s.ServiceTotal += i
	s.ServiceMin = MinButNotZero(s.ServiceMin, i)
	s.ServiceMax = Max(s.ServiceMax, i)
}

func (s *ServiceInfo) UpdateAvg(i []float64) {
	s.ServiceAvg = Clamp(0, Round(Average(i)), 100)
}

func (s *ServiceInfo) UpdateDetails(sc []rancher.APIService) {
	for _, secret := range sc {
		if secret.Labels[projectLabel] == "" {
			s.NoProject++
		}
	}
}

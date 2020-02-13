package collector

type PodInfo struct {
	PodsMin   int `json:"pods_min"`
	PodsMax   int `json:"pods_max"`
	PodsTotal int `json:"pods_total"`
	UtilMin   int `json:"util_min"`
	UtilAvg   int `json:"util_avg"`
	UtilMax   int `json:"util_max"`
}

func (p *PodInfo) Update(total, util int) {
	p.PodsMin = MinButNotZero(p.PodsMin, total)
	p.PodsMax = Max(p.PodsMax, total)
	p.PodsTotal += total
	p.UtilMin = MinButNotZero(p.UtilMin, util)
	p.UtilMax = Max(p.UtilMax, util)
}

func (p *PodInfo) UpdateAvg(i []float64) {
	p.UtilAvg = Clamp(0, Round(Average(i)), 100)
}

type PodData struct {
	PodMin   int `json:"min"`
	PodMax   int `json:"max"`
	PodTotal int `json:"total"`
	PodAvg   int `json:"avg"`
}

func (w *PodData) Update(i int) {
	w.PodTotal += i
	w.PodMin = MinButNotZero(w.PodMin, i)
	w.PodMax = Max(w.PodMax, i)
}

func (w *PodData) UpdateAvg(i []float64) {
	w.PodAvg = Clamp(0, Round(Average(i)), 100)
}

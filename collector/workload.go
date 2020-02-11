package collector

type WorkloadInfo struct {
	WorkloadMin   int `json:"min"`
	WorkloadMax   int `json:"max"`
	WorkloadTotal int `json:"total"`
	WorkloadAvg   int `json:"avg"`
}

func (w *WorkloadInfo) Update(i int) {
	w.WorkloadTotal += i
	w.WorkloadMin = MinButNotZero(w.WorkloadMin, i)
	w.WorkloadMax = Max(w.WorkloadMax, i)
}

func (w *WorkloadInfo) UpdateAvg(i []float64) {
	w.WorkloadAvg = Clamp(0, Round(Average(i)), 100)
}

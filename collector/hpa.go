package collector

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

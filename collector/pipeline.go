package collector

type PipelineInfo struct {
	Enabled        int        `json:"enabled"` // 1 if user has any # pipeline provider enabled
	SourceProvider LabelCount `json:"source"`
	TotalPipelines int        `json:"total"`
}

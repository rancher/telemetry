package collector_test

import (
	"testing"

	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func checkLabelCount(t *testing.T, expected collector.LabelCount, actual collector.LabelCount) {
	assert.Equal(t, len(expected), len(actual))
	for key, val := range expected {
		valActual, found := actual[key]
		assert.True(t, found)
		assert.Equal(t, val, valActual)
	}
}

func checkCPUInfo(t *testing.T, expected collector.CpuInfo, actual collector.CpuInfo) {
	assert.Equal(t, expected.CoresTotal, actual.CoresTotal)
	assert.Equal(t, expected.CoresMax, actual.CoresMax)
	assert.Equal(t, expected.CoresMin, actual.CoresMin)
	assert.Equal(t, expected.UtilAvg, actual.UtilAvg)
	assert.Equal(t, expected.UtilMax, actual.UtilMax)
	assert.Equal(t, expected.UtilMin, actual.UtilMin)
}

func checkMemoryInfo(t *testing.T, expected collector.MemoryInfo, actual collector.MemoryInfo) {
	assert.Equal(t, expected.TotalMb, actual.TotalMb)
	assert.Equal(t, expected.MaxMb, actual.MaxMb)
	assert.Equal(t, expected.MinMb, actual.MinMb)
	assert.Equal(t, expected.UtilAvg, actual.UtilAvg)
	assert.Equal(t, expected.UtilMax, actual.UtilMax)
	assert.Equal(t, expected.UtilMin, actual.UtilMin)
}

func checkPodInfo(t *testing.T, expected collector.PodInfo, actual collector.PodInfo) {
	assert.Equal(t, expected.PodsTotal, actual.PodsTotal)
	assert.Equal(t, expected.PodsMax, actual.PodsMax)
	assert.Equal(t, expected.PodsMin, actual.PodsMin)
	assert.Equal(t, expected.UtilAvg, actual.UtilAvg)
	assert.Equal(t, expected.UtilMax, actual.UtilMax)
	assert.Equal(t, expected.UtilMin, actual.UtilMin)
}

func checkNsInfo(t *testing.T, expected collector.NsInfo, actual collector.NsInfo) {
	assert.Equal(t, expected.NsTotal, actual.NsTotal)
	assert.Equal(t, expected.NsMax, actual.NsMax)
	assert.Equal(t, expected.NsMin, actual.NsMin)
	assert.Equal(t, expected.NsAvg, actual.NsAvg)
	assert.Equal(t, expected.FromCatalog, actual.FromCatalog)
	assert.Equal(t, expected.NoProject, actual.NoProject)
}

func checkWorkloadInfo(t *testing.T, expected collector.WorkloadInfo, actual collector.WorkloadInfo) {
	assert.Equal(t, expected.WorkloadTotal, actual.WorkloadTotal)
	assert.Equal(t, expected.WorkloadMax, actual.WorkloadMax)
	assert.Equal(t, expected.WorkloadMin, actual.WorkloadMin)
	assert.Equal(t, expected.WorkloadAvg, actual.WorkloadAvg)
}

func checkPipelineInfo(t *testing.T, expected collector.PipelineInfo, actual collector.PipelineInfo) {
	assert.Equal(t, expected.Enabled, actual.Enabled)
	assert.Equal(t, expected.SourceProvider, actual.SourceProvider)
	assert.Equal(t, expected.TotalPipelines, actual.TotalPipelines)
}

func checkHPAInfo(t *testing.T, expected collector.HPAInfo, actual collector.HPAInfo) {
	assert.Equal(t, expected.Total, actual.Total)
	assert.Equal(t, expected.Max, actual.Max)
	assert.Equal(t, expected.Min, actual.Min)
	assert.Equal(t, expected.Avg, actual.Avg)
}

func checkPodData(t *testing.T, expected collector.PodData, actual collector.PodData) {
	assert.Equal(t, expected.PodTotal, actual.PodTotal)
	assert.Equal(t, expected.PodMax, actual.PodMax)
	assert.Equal(t, expected.PodMin, actual.PodMin)
	assert.Equal(t, expected.PodAvg, actual.PodAvg)
}

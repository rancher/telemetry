package collector_test

import (
	"testing"

	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func TestUtilsGetMemBadUnit(t *testing.T) {
	// no unit means bytes
	val := collector.GetMem("1234", "")
	assert.Equal(t, int64(1234), val)
}

func TestUtilsGetCPUSuffixM(t *testing.T) {
	val := collector.GetCPU("2000m")
	assert.Equal(t, 2, val)
}

func TestUtilsGetRawInt64ItemEqualSuffix(t *testing.T) {
	val := collector.GetRawInt64("m", "m")
	assert.Equal(t, int64(0), val)
}

func TestUtilsGetRawInt64ItemEmpty(t *testing.T) {
	val := collector.GetRawInt64("", "m")
	assert.Equal(t, int64(0), val)
}

func TestUtilsGetRawInt64NotInt(t *testing.T) {
	val := collector.GetRawInt64("this_is_not_an_int", "m")
	assert.Equal(t, int64(0), val)
}

func TestUtilsGetRawIntItemEqualSuffix(t *testing.T) {
	val := collector.GetRawInt("m", "m")
	assert.Equal(t, 0, val)
}

func TestUtilsGetRawIntItemEmpty(t *testing.T) {
	val := collector.GetRawInt("", "m")
	assert.Equal(t, 0, val)
}

func TestUtilsGetRawIntNotInt(t *testing.T) {
	val := collector.GetRawInt("this_is_not_an_int", "m")
	assert.Equal(t, 0, val)
}

func TestUtilsFromCatalogTrue(t *testing.T) {
	val := collector.FromCatalog("catalog://TRUE")
	assert.True(t, val)
}

func TestUtilsFromCatalogFalse(t *testing.T) {
	val := collector.FromCatalog("BLAH_BLAH_BLAH_TEST")
	assert.False(t, val)
}

func TestUtilsMinSecondTerm(t *testing.T) {
	val := collector.Min(3, 1)
	assert.Equal(t, 1, val)
}

func TestUtilsMinFirstTerm(t *testing.T) {
	val := collector.Min(2, 3)
	assert.Equal(t, 2, val)
}

func TestUtilsLabelCountIncrement(t *testing.T) {
	label := collector.LabelCount{}
	label.Increment("test1")
	label.Increment("test1")
	label.Increment("test1")
	label.Increment("test2")
	label.Increment("test2")
	label.Increment("test3")
	label.Increment("")
	label.Increment("")
	checkLabelCount(t, collector.LabelCount{"test1": 3, "test2": 2, "test3": 1, "(unknown)": 2}, label)
}

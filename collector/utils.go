package collector

import (
	"strconv"
	"strings"

	"github.com/rancher/norman/clientbase"
	norman "github.com/rancher/norman/types"
	log "github.com/sirupsen/logrus"
)

const (
	catalogProto = "catalog://"
	searchLimit  = "-1"
)

type CpuInfo struct {
	CoresMin   int `json:"cores_min"`
	CoresMax   int `json:"cores_max"`
	CoresTotal int `json:"cores_total"`
	UtilMin    int `json:"util_min"`
	UtilAvg    int `json:"util_avg"`
	UtilMax    int `json:"util_max"`
}

func (c *CpuInfo) Update(total, util int) {
	c.CoresMin = MinButNotZero(c.CoresMin, total)
	c.CoresMax = Max(c.CoresMin, total)
	c.CoresTotal += total
	c.UtilMin = MinButNotZero(c.UtilMin, util)
	c.UtilMax = Max(c.UtilMax, util)
}

func (c *CpuInfo) UpdateAvg(i []float64) {
	c.UtilAvg = Clamp(0, Round(Average(i)), 100)
}

type MemoryInfo struct {
	MinMb   int `json:"mb_min"`
	MaxMb   int `json:"mb_max"`
	TotalMb int `json:"mb_total"`
	UtilMin int `json:"util_min"`
	UtilAvg int `json:"util_avg"`
	UtilMax int `json:"util_max"`
}

func (m *MemoryInfo) Update(total, util int) {
	m.MinMb = MinButNotZero(m.MinMb, total)
	m.MaxMb = Max(m.MaxMb, total)
	m.TotalMb += total
	m.UtilMin = MinButNotZero(m.UtilMin, util)
	m.UtilMax = Max(m.UtilMax, util)
}

func (m *MemoryInfo) UpdateAvg(i []float64) {
	m.UtilAvg = Clamp(0, Round(Average(i)), 100)
}

func GetMemMb(item string) int {
	return int(GetMem(item, "Mi"))
}

func GetMem(item, unit string) int64 {
	units := map[string]int64{
		"Ki": 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"Mi": 1024 * 1024,
		"G":  1000 * 1000 * 1000,
		"Gi": 1024 * 1024 * 1024,
	}

	outUnit := units[unit]
	if unit == "" || outUnit == 0 {
		log.Debugf("GetMem using default output unit [byte]")
		outUnit = 1
	}

	for key, value := range units {
		if strings.HasSuffix(item, key) {
			return GetRawInt64(item, key) * value / outUnit
		}
	}

	return GetRawInt64(item, "") / outUnit
}

func GetCPU(item string) int {
	key := "m"
	value := float64(1000)

	if strings.HasSuffix(item, key) {
		utilFloat := float64(GetRawInt(item, key)) / value
		return Round(utilFloat)
	}

	return GetRawInt(item, "")
}

func GetRawInt64(item, sep string) int64 {
	if item == "" || item == sep {
		return int64(0)
	}

	toConv := item
	if sep != "" {
		toConv = strings.Replace(item, sep, "", 1)
	}

	result, err := strconv.ParseInt(toConv, 10, 64)
	// If error or result is negative returning 0
	if err != nil || result < 0 {
		log.Debugf("Error converting string to int64 [%s] %s", toConv, err)
		return int64(0)
	}

	return result
}

func GetRawInt(item, sep string) int {
	if item == "" || item == sep {
		return 0
	}

	toConv := item
	if sep != "" {
		toConv = strings.Replace(item, sep, "", 1)
	}

	result, err := strconv.Atoi(toConv)
	// If error or result is negative returning 0
	if err != nil || result < 0 {
		log.Debugf("Error converting string to int [%s] %s", toConv, err)
		return 0
	}

	return result
}

func FromCatalog(s string) bool {
	return strings.Contains(s, catalogProto)
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MinButNotZero(x, y int) int {
	if x == 0 || x > y {
		return y
	}
	return x
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Average(x []float64) float64 {
	var sum float64
	num := len(x)

	if num == 0 {
		return 0.0
	}

	for _, value := range x {
		sum += value
	}

	return sum / float64(num)
}

func Round(f float64) int {
	return int(f + 0.5)
}

func Clamp(min, x, max int) int {
	return Max(min, Min(x, max))
}

func NonRemoved() norman.ListOpts {
	filters := make(map[string]interface{})
	filters["state_ne"] = "removed"
	filters["limit"] = searchLimit

	out := norman.ListOpts{
		Filters: filters,
	}

	return out
}

type LabelCount map[string]int

func (m *LabelCount) Increment(k string) {
	if len(k) == 0 {
		k = "(unknown)"
	}

	cur, ok := (*m)[k]
	if ok {
		(*m)[k] = cur + 1
	} else {
		(*m)[k] = 1
	}
}

func IsNotFound(err error) bool {
	return clientbase.IsNotFound(err)
}

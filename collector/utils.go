package collector

import (
	"strconv"
	"strings"

	norman "github.com/rancher/norman/types"
	rancher "github.com/rancher/types/client/management/v3"
	log "github.com/sirupsen/logrus"
)

const catalogProto = "catalog://"

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

func GetNodeTemplate(cli *rancher.Client, id string) *rancher.NodeTemplate {
	if id == "" {
		log.Debugf("nodeTemplate id is empty")
		return nil
	}

	base := cli.Opts.URL
	url := base + "/nodeTemplates/" + id

	mTemplate := &rancher.NodeTemplate{}
	version := "nodeTemplate"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := cli.GetLink(resource, version, mTemplate)

	if mTemplate == nil || mTemplate.Type != "nodeTemplate" {
		log.Debugf("nodeTemplate not found [%s]", resource.Links[version])
		return nil
	}
	if err != nil {
		log.Debugf("Error getting nodeTemplate [%s] %s", resource.Links[version], err)
		return nil
	}

	return mTemplate
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

func GetRawInt64(item, sep string) int64 {
	if item == "" {
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
	filters["limit"] = "-1"

	out := norman.ListOpts{
		Filters: filters,
	}

	return out
}

func GetSetting(client *rancher.Client, key string) (string, bool) {
	return GetSettingByCollection(GetSettingCollection(client), key)
}

func GetSettingCollection(client *rancher.Client) *rancher.SettingCollection {
	opts := NonRemoved()
	opts.Filters["all"] = "true"

	settings, err := client.Setting.List(&opts)
	if err != nil {
		log.Errorf("GetSettingsCollection: Error: %s", err)
		return nil
	}

	if settings == nil || settings.Type != "collection" || len(settings.Data) == 0 {
		log.Debugf("Settings collection is empty")
		return nil
	}

	return settings
}

func GetSettingByCollection(settings *rancher.SettingCollection, key string) (string, bool) {
	if settings == nil || key == "" {
		return "", false
	}
	for _, setting := range settings.Data {
		if setting.ID == key {
			if setting.Value == "" {
				log.Debugf("GetSetting(%s): Not Set", key)
			} else {
				log.Debugf("GetSetting(%s) = %s", key, setting.Value)
			}
			return setting.Value, true
		}
	}
	return "", false
}

func SetSetting(client *rancher.Client, key string, value string) error {
	setting, err := client.Setting.ByID(key)
	if err == nil {
		_, err = client.Setting.Update(setting, map[string]interface{}{"value": value})
		if err == nil {
			log.Debugf("UpdateSetting(%s,%s)", key, value)
		} else {
			log.Debugf("UpdateSetting(%s,%s): Error: %s", key, value, err)
		}
		return err
	}

	setting, err = client.Setting.Create(&rancher.Setting{
		Name:  key,
		Value: value,
	})

	if err == nil {
		log.Debugf("CreateSetting(%s,%s)", key, value)
	} else {
		log.Debugf("CreateSetting(%s,%s): Error: %s", key, value, err)
	}
	return err
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

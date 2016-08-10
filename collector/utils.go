package collector

import (
	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/client"
)

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

func IncrementMap(m *map[string]int, k string) {
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

func NonRemoved() rancher.ListOpts {
	filters := make(map[string]interface{})
	filters["state_ne"] = "removed"

	out := rancher.ListOpts{
		Filters: filters,
	}

	return out
}

func GetSetting(client *rancher.RancherClient, key string) (string, bool) {
	setting, err := client.Setting.ById(key)
	if err != nil {
		log.Errorf("Failed to get Setting key=%s, err=%s", key, err)
		return "", false
	}

	return setting.Value, true
}

package collector

import (
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	log "github.com/sirupsen/logrus"
)

func SetSetting(client *rancher.Client, key string, value string) error {
	setting, err := client.Setting.ByID(key)
	if err != nil {
		if IsNotFound(err) {
			setting, err = client.Setting.Create(&rancher.Setting{
				Name:  key,
				Value: value,
			})
			if err == nil {
				log.Debugf("CreateSetting(%s,%s)", key, value)
			} else {
				log.Debugf("CreateSetting(%s,%s): Error: %s", key, value, err)
			}
		} else {
			log.Debugf("Failed to get setting %s err=%s", key, err)
		}
		return err
	}
	_, err = client.Setting.Update(setting, map[string]interface{}{"value": value})
	if err == nil {
		log.Debugf("UpdateSetting(%s,%s)", key, value)
	} else {
		log.Debugf("UpdateSetting(%s,%s): Error: %s", key, value, err)
	}
	return err
}

package collector

import (
	norman "github.com/rancher/norman/types"
	client "github.com/rancher/types/client/project/v3"
	log "github.com/sirupsen/logrus"
)

func GetAppsCollection(c *CollectorOpts, url string) *client.AppCollection {
	if url == "" {
		log.Debugf("App collection link is empty.")
		return nil
	}

	appCollection := &client.AppCollection{}
	version := "apps"

	resource := norman.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.Client.GetLink(resource, version, appCollection)
	if err != nil {
		log.Errorf("Error getting app collection [%s] %s", resource.Links[version], err)
		return nil
	}

	if appCollection == nil || appCollection.Type != "collection" || len(appCollection.Data) == 0 {
		log.Debugf("App collection is empty [%s]", resource.Links[version])
		return nil
	}

	return appCollection
}

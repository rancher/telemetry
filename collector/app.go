package collector

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	AppTemplateExternalIDPrefix = "catalog://?"
	AppCatalogLibrary           = "library"
	AppCatalogSystemLibrary     = "system-library"
)

var (
	AppRancherCatalogs  = []string{AppCatalogLibrary, AppCatalogSystemLibrary}
	AppGetProjectClient = GetProjectClient
)

type AppTemplate struct {
	State string                 `json:"state"`
	Apps  map[string]*LabelCount `json:"apps"`
}

type App struct {
	Total    int                     `json:"total"`
	Active   int                     `json:"active"`
	Catalogs map[string]*AppTemplate `json:"rancheCatalogs"`
}

func (a App) RecordKey() string {
	return "app"
}

func (a App) Collect(c *CollectorOpts) interface{} {
	log.Debug("Collecting Apps")
	opts := NonRemoved()
	opts.Filters["all"] = "true"

	nonRemoved := NonRemoved()

	a.Catalogs = map[string]*AppTemplate{}

	for _, catalog := range AppRancherCatalogs {
		state, err := GetAppCatalogState(c, catalog)
		if err != nil {
			log.Errorf("Failed to get Catalog ID %s err=%s", catalog, err)
			return nil
		}
		a.Catalogs[catalog] = &AppTemplate{
			State: state,
			Apps:  map[string]*LabelCount{},
		}
	}

	log.Debug("  Collecting Projects")
	projectList, err := c.Client.Project.ListAll(&opts)
	if err != nil {
		log.Errorf("Failed to get Projects err=%s", err)
		return nil
	}
	log.Debugf("  Found %d Projects", len(projectList.Data))

	for _, project := range projectList.Data {
		projectClient, err := AppGetProjectClient(c, project.ID)
		if err != nil {
			log.Errorf("Failed to get project client ID %s err=%s", project.ID, err)
			continue
		}

		log.Debugf("  Collecting Apps")
		appsCollection, err := projectClient.App.ListAll(&nonRemoved)
		if err != nil {
			log.Errorf("Failed to get Apps for project %s err=%s", project.ID, err)
		} else {
			log.Debugf("  Found %d Apps", len(appsCollection.Data))
			for _, app := range appsCollection.Data {
				externalID, err := SplitAppExternalID(app.ExternalID)
				if err != nil {
					log.Errorf("Failed to split App External ID %s err=%s", app.ExternalID, err)
					continue
				}
				if a.Catalogs[externalID["catalog"]] == nil {
					continue
				}
				a.Total++
				if app.State == "active" {
					a.Active++
				}

				if a.Catalogs[externalID["catalog"]].Apps[externalID["template"]] == nil {
					a.Catalogs[externalID["catalog"]].Apps[externalID["template"]] = &LabelCount{}
				}
				a.Catalogs[externalID["catalog"]].Apps[externalID["template"]].Increment(externalID["version"])
			}
		}
	}

	return a
}

func SplitAppExternalID(externalID string) (map[string]string, error) {
	//Global catalog url: catalog://?catalog=demo&template=test&version=1.23.0
	//Cluster catalog url: catalog://?catalog=c-XXXXX/test&type=clusterCatalog&template=test&version=1.23.0
	//Project catalog url: catalog://?catalog=p-XXXXX/test&type=projectCatalog&template=test&version=1.23.0

	val, err := url.Parse(externalID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	out["catalog"] = val.Query().Get("catalog")
	out["type"] = val.Query().Get("type")
	//Setting proper catalog name if type is clusterCatalog or projectCatalog
	if out["type"] == "clusterCatalog" || out["type"] == "projectCatalog" {
		out["catalog"] = strings.Replace(out["catalog"], "/", ":", -1)
	}
	out["template"] = val.Query().Get("template")
	out["version"] = val.Query().Get("version")

	if out["catalog"] == "" || out["template"] == "" || out["version"] == "" {
		return nil, fmt.Errorf("Bad External ID format")
	}

	return out, nil
}

func GetAppCatalogState(c *CollectorOpts, id string) (string, error) {
	catalog, err := c.Client.Catalog.ByID(id)
	if err != nil {
		if IsNotFound(err) {
			return "disabled", nil
		}
		return "", err
	}

	return catalog.State, nil
}

func init() {
	Register(App{})
}

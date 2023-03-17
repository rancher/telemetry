package collector_test

import (
	"fmt"
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewClientForTestingApp(data string) (*rancher.Client, error) {
	client := &rancher.Client{}
	client.Project = NewProjectOperationsMock(data)
	client.Opts = NewTestClientOpts()
	return client, nil
}

func TestAppBasic(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0",
			"state": "active"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 1, resApp.Total)
	assert.Equal(t, 1, resApp.Active)
	appTemplate, ok := resApp.Catalogs["library"]
	assert.True(t, ok)
	assert.NotNil(t, appTemplate)
	assert.Equal(t, "active", appTemplate.State)
	appCount, ok := appTemplate.Apps["app1"]
	assert.True(t, ok)
	assert.Equal(t, collector.LabelCount{"1.23.0": 1}, *appCount)
}

func TestAppMoreApps(t *testing.T) {
	c := collector.App{}
	appClientProject1 := &rancherProject.Client{}
	appClientProject1.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0",
			"state": "active"
		},
		{
			"id": "app2",
			"type": "app",
			"externalId": "catalog://?catalog=system-library&type=projectCatalog&template=app2&version=1.55.0",
			"state": "active"
		}
		]`)
	collector.ProjectClients["project1"] = appClientProject1

	appClientProject2 := &rancherProject.Client{}
	appClientProject2.App = NewAppOperationsMock(
		`[
		{
			"id": "app3",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app3&version=1.33.0"
		},
		{
			"id": "app4",
			"type": "app",
			"externalId": "catalog://?catalog=ignored&type=projectCatalog&template=app4&version=1.44.0"
		}
		]`)
	collector.ProjectClients["project2"] = appClientProject2

	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			},
			{
				"id": "project2",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 3, resApp.Total)
	assert.Equal(t, 2, resApp.Active)
	// library should have apps 1 and 3
	appTemplate, ok := resApp.Catalogs["library"]
	assert.True(t, ok)
	assert.NotNil(t, appTemplate)
	assert.Equal(t, "active", appTemplate.State)
	appCount1, ok := appTemplate.Apps["app1"]
	assert.True(t, ok)
	assert.Equal(t, collector.LabelCount{"1.23.0": 1}, *appCount1)
	appCount2, ok := appTemplate.Apps["app3"]
	assert.True(t, ok)
	assert.Equal(t, collector.LabelCount{"1.33.0": 1}, *appCount2)

	appTemplate, ok = resApp.Catalogs["system-library"]
	assert.True(t, ok)
	assert.NotNil(t, appTemplate)
	assert.Equal(t, "active", appTemplate.State)
	appCount1, ok = appTemplate.Apps["app2"]
	assert.True(t, ok)
	assert.Equal(t, collector.LabelCount{"1.55.0": 1}, *appCount1)
}

func TestAppGetAppCatalogStateError(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock("FAIL_BY_ID")
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestAppGetAppCatalogStateErrorNotFound(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock("FAIL_BY_ID_NOT_FOUND")
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 1, resApp.Total)
	appTemplate, ok := resApp.Catalogs["library"]
	assert.True(t, ok)
	assert.NotNil(t, appTemplate)
	assert.Equal(t, "disabled", appTemplate.State)
	appCount, ok := appTemplate.Apps["app1"]
	assert.True(t, ok)
	assert.Equal(t, collector.LabelCount{"1.23.0": 1}, *appCount)
}

func TestAppListAllAppError(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock("FAIL_LIST_ALL")
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, 0, resApp.Total)
}

func TestAppListAllProjectError(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp("FAIL_LIST_ALL")
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestAppAppGetProjectClientError(t *testing.T) {
	prevFunc := collector.AppGetProjectClient
	defer func() { collector.AppGetProjectClient = prevFunc }()
	collector.AppGetProjectClient = func(c *collector.CollectorOpts, id string) (*rancherProject.Client, error) {
		return nil, fmt.Errorf("ERROR: GetProjectClient Test")
	}
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 0, resApp.Total)
}

func TestAppErrorParsingCatalogURL(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "postgres://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 0, resApp.Total)
}

func TestAppCatalogEmptyVersion(t *testing.T) {
	c := collector.App{}
	appClient := &rancherProject.Client{}
	appClient.App = NewAppOperationsMock(
		`[
		{
			"id": "app1",
			"type": "app",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1",
			"state": "active"
		}
		]`)
	collector.ProjectClients["project1"] = appClient
	client, _ := NewClientForTestingApp(
		`[
			{
				"id": "project1",
				"type": "project"
			}
		]`)
	client.Catalog = NewCatalogOperationsMock(`{"id": "catalog1", "state": "active" }`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resApp := res.(collector.App)
	assert.Equal(t, "app", resApp.RecordKey())
	assert.Equal(t, 0, resApp.Total)
	assert.Equal(t, 0, resApp.Active)
}

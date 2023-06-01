package collector_test

import (
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewMCAClientMock(mca string, catalogState string,
	templateVersion string, globalDnsProvider string,
	globalDns string) *rancher.Client {
	client := &rancher.Client{}
	client.MultiClusterApp = NewMultiClusterAppOperationsMock(mca)
	client.Catalog = NewCatalogOperationsMock(catalogState)
	client.TemplateVersion = NewTemplateVersionOperationsMock(templateVersion)
	client.GlobalDnsProvider = NewGlobalDnsProviderOperationsMock(globalDnsProvider)
	client.GlobalDns = NewGlobalDnsOperationsMock(globalDns)
	client.Opts = NewTestClientOpts()
	return client
}

func TestMCABasic(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 1, resMCA.Total)
	assert.Equal(t, 1, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 2, resMCA.TargetMax)
	assert.Equal(t, 2, resMCA.TargetMin)
	assert.Equal(t, 2, resMCA.TargetTotal)
	assert.Equal(t, 2.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.True(t, okLibrary)
	assert.NotNil(t, libraryCatalogs)
	assert.Equal(t, "active", libraryCatalogs.State)
	app1, okApp1 := libraryCatalogs.Apps["app1"]
	assert.True(t, okApp1)
	assert.NotNil(t, app1)
	checkLabelCount(t, collector.LabelCount{"1.23.0": 1}, *app1)
}

func TestMCAGetCatalogStateError(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`FAIL_BY_ID`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestMCATemplateVersionError(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		"FAIL_BY_ID",
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 1, resMCA.Total)
	assert.Equal(t, 1, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 2, resMCA.TargetMax)
	assert.Equal(t, 2, resMCA.TargetMin)
	assert.Equal(t, 2, resMCA.TargetTotal)
	assert.Equal(t, 2.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.True(t, okLibrary)
	assert.NotNil(t, libraryCatalogs)
	assert.Equal(t, "active", libraryCatalogs.State)
	app1, okApp1 := libraryCatalogs.Apps["app1"]
	assert.False(t, okApp1)
	assert.Nil(t, app1)
}

func TestMCACatalogNotValid(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "catalog://?catalog=CATALOG_TEST&type=clusterCatalog&template=app1&version=1.23.0"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 1, resMCA.Total)
	assert.Equal(t, 1, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 2, resMCA.TargetMax)
	assert.Equal(t, 2, resMCA.TargetMin)
	assert.Equal(t, 2, resMCA.TargetTotal)
	assert.Equal(t, 2.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.True(t, okLibrary)
	assert.NotNil(t, libraryCatalogs)
	assert.Equal(t, "active", libraryCatalogs.State)
	app1, okApp1 := libraryCatalogs.Apps["app1"]
	assert.False(t, okApp1)
	assert.Nil(t, app1)
}

func TestMCABadCatalogURL(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "postgres://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 1, resMCA.Total)
	assert.Equal(t, 1, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 2, resMCA.TargetMax)
	assert.Equal(t, 2, resMCA.TargetMin)
	assert.Equal(t, 2, resMCA.TargetTotal)
	assert.Equal(t, 2.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.True(t, okLibrary)
	assert.NotNil(t, libraryCatalogs)
	assert.Equal(t, "active", libraryCatalogs.State)
	app1, okApp1 := libraryCatalogs.Apps["app1"]
	assert.False(t, okApp1)
	assert.Nil(t, app1)
}

func TestMCACatalogEmptyCatalog(t *testing.T) {
	client := NewMCAClientMock(
		`[
			{
				"id": "mca1",
				"type": "mca",
				"state": "active",
				"targets" : [
					{
						"appId" : "app1"
					},
					{
						"appId" : "app2"
					}
				 ],
				"templateVersionId" : "versionId"
			}
		]`,
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "catalog://?type=clusterCatalog&template=app1&version=1.23.0"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 1, resMCA.Total)
	assert.Equal(t, 1, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 2, resMCA.TargetMax)
	assert.Equal(t, 2, resMCA.TargetMin)
	assert.Equal(t, 2, resMCA.TargetTotal)
	assert.Equal(t, 2.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.True(t, okLibrary)
	assert.NotNil(t, libraryCatalogs)
	assert.Equal(t, "active", libraryCatalogs.State)
	app1, okApp1 := libraryCatalogs.Apps["app1"]
	assert.False(t, okApp1)
	assert.Nil(t, app1)
}

func TestMCAListAllError(t *testing.T) {
	client := NewMCAClientMock(
		"FAIL_LIST_ALL",
		`
		{
			"id" : "catalogState1",
			"type": "state",
			"state": "active"
		}
		`,
		`
		{
			"id" : "templateVersion1",
			"type": "templateVersion",
			"externalId": "catalog://?catalog=library&type=clusterCatalog&template=app1&version=1.23.0"
		}
		`,
		`[
			{
				"id" : "globalDnsProvider1",
				"type" : "globalDnsProvider"
			},
			{
				"id" : "globalDnsProvider2",
				"type" : "globalDnsProvider"
			}
		]`,
		`[
			{
				"id" : "globalDns1",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns2",
				"type" : "globalDns"
			},
			{
				"id" : "globalDns3s",
				"type" : "globalDns"
			}
		]`)
	c := collector.MultiClusterApp{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resMCA := res.(collector.MultiClusterApp)
	assert.NotNil(t, resMCA)
	assert.Equal(t, "mca", resMCA.RecordKey())
	assert.Equal(t, 0, resMCA.Total)
	assert.Equal(t, 0, resMCA.Active)
	assert.Equal(t, 2, resMCA.DnsProviders)
	assert.Equal(t, 3, resMCA.DnsEntries)
	assert.Equal(t, 0, resMCA.TargetMax)
	assert.Equal(t, 0, resMCA.TargetMin)
	assert.Equal(t, 0, resMCA.TargetTotal)
	assert.Equal(t, 0.0, resMCA.TargetAvg)
	libraryCatalogs, okLibrary := resMCA.Catalogs["library"]
	assert.False(t, okLibrary)
	assert.Nil(t, libraryCatalogs)
}

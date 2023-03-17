package collector_test

import (
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func TestClusterTemplateBasic(t *testing.T) {
	client := &rancher.Client{}
	client.ClusterTemplate = NewClusterTemplateOperationsMock(`[
		{
			"id": "clustertemplate1",
	     	"type": "clustertemplate"
		},
		{
			"id": "clustertemplate2",
	     	"type": "clustertemplate"
		}
	]`)
	client.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`[
		{
			"id": "clustertemplaterev1",
	     	"type": "clustertemplaterevision"
		},
		{
			"id": "clustertemplaterev2",
	     	"type": "clustertemplaterevision"
		}
	]`)
	client.Setting = NewSettingOperationsMock(`
	{
		"id": "setting1",
		"type": "setting",
		"default": "DEFAULT_VALUE_TEST",
		"value": "VALUE_TEST"
	}`, nil)
	client.Opts = NewTestClientOpts()
	opts := &collector.CollectorOpts{
		Client: client,
	}
	c := collector.ClusterTemplate{}
	res := c.Collect(opts)
	resApp := res.(collector.ClusterTemplate)
	assert.Equal(t, 2, resApp.TotalClusterTemplates)
	assert.Equal(t, 2, resApp.TotalTemplateRevisions)
	assert.Equal(t, "VALUE_TEST", resApp.Enforcement)
	assert.Equal(t, "clustertemplate", resApp.RecordKey())
}

func TestClusterTemplateBasicDefaultValue(t *testing.T) {
	client := &rancher.Client{}
	client.ClusterTemplate = NewClusterTemplateOperationsMock(`[
		{
			"id": "clustertemplate1",
	     	"type": "clustertemplate"
		}
	]`)
	client.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`[
		{
			"id": "clustertemplaterev1",
	     	"type": "clustertemplaterevision"
		}
	]`)
	client.Setting = NewSettingOperationsMock(`
	{
		"id": "setting1",
		"type": "setting",
		"default": "DEFAULT_VALUE_TEST"
	}`, nil)
	client.Opts = NewTestClientOpts()
	opts := &collector.CollectorOpts{
		Client: client,
	}
	c := collector.ClusterTemplate{}
	res := c.Collect(opts)
	resApp := res.(collector.ClusterTemplate)
	assert.Equal(t, 1, resApp.TotalClusterTemplates)
	assert.Equal(t, 1, resApp.TotalTemplateRevisions)
	assert.Equal(t, "DEFAULT_VALUE_TEST", resApp.Enforcement)
}

func TestClusterTemplateClusterTemplateListAllError(t *testing.T) {
	client := &rancher.Client{}
	client.ClusterTemplate = NewClusterTemplateOperationsMock(`FAIL_LIST_ALL`)
	client.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`[
		{
			"id": "clustertemplaterev1",
	     	"type": "clustertemplaterevision"
		}
	]`)
	client.Setting = NewSettingOperationsMock(`
	{
		"id": "setting1",
		"type": "setting",
		"default": "DEFAULT_VALUE_TEST"
	}`, nil)
	client.Opts = NewTestClientOpts()
	opts := &collector.CollectorOpts{
		Client: client,
	}
	c := collector.ClusterTemplate{}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestClusterTemplateClusterTemplateRevListAllError(t *testing.T) {
	client := &rancher.Client{}
	client.ClusterTemplate = NewClusterTemplateOperationsMock(`[
		{
			"id": "clustertemplate1",
	     	"type": "clustertemplate"
		}
	]`)
	client.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`FAIL_LIST_ALL`)
	client.Setting = NewSettingOperationsMock(`
	{
		"id": "setting1",
		"type": "setting",
		"default": "DEFAULT_VALUE_TEST"
	}`, nil)
	client.Opts = NewTestClientOpts()
	opts := &collector.CollectorOpts{
		Client: client,
	}
	c := collector.ClusterTemplate{}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestClusterTemplateSettingByIDError(t *testing.T) {
	client := &rancher.Client{}
	client.ClusterTemplate = NewClusterTemplateOperationsMock(`[
		{
			"id": "clustertemplate1",
	     	"type": "clustertemplate"
		}
	]`)
	client.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`[
		{
			"id": "clustertemplaterev1",
	     	"type": "clustertemplaterevision"
		}
	]`)
	client.Setting = NewSettingOperationsMock(`FAIL_BY_ID`, nil)
	client.Opts = NewTestClientOpts()
	opts := &collector.CollectorOpts{
		Client: client,
	}
	c := collector.ClusterTemplate{}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

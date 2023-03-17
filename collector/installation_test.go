package collector_test

import (
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewInstallationClientMock(authConfig string, user string,
	nodeDriver string, kontainerDriver string,
	cluster string, settings map[string]string) *rancher.Client {
	client := &rancher.Client{}
	client.AuthConfig = NewAuthConfigOperationsMock(authConfig)
	client.User = NewUserOperationsMock(user)
	client.NodeDriver = NewNodeDriverOperationsMock(nodeDriver)
	client.KontainerDriver = NewKontainerDriverOperationsMock(kontainerDriver)
	client.Cluster = NewClusterOperationsMock(cluster)
	client.Setting = NewSettingOperationsMock("{}", settings)
	client.Opts = NewTestClientOpts()
	return client
}

func TestInstallationBasic(t *testing.T) {
	settings := map[string]string{
		"server-version":     "SERVER_VERSION",
		"telemetry-uid":      "TELEMETRY_UID",
		"ui-default-landing": "UI-DEFAULT-LANDING",
	}
	client := NewInstallationClientMock(`
	[
		{
			"id" : "authConfig1",
			"enabled": true,
			"type": "TEST1Config"
		}
	]`,
		`[
		{
			"id" : "user1",
			"type": "user",
			"principalIds" : ["id1://test", "id2://test"]
		}
	]`,
		`[
		{
			"id" : "nodeDriver1",
			"type" : "nodeDriver",
			"active" : true,
			"name" : "NODE_DRIVER_TEST"
		}
	]`,
		`[
		{
			"id" : "kontainerDriver1",
			"type" : "kontainerDriver",
			"active" : true,
			"name" : "KONTAINER_DRIVER_TEST"
		}
	]`,
		`[
			{
				"id" : "cluster1",
				"type": "cluster",
				"internal": true
			}
	]`,
		settings)

	c := collector.Installation{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resInstallation := res.(collector.Installation)
	assert.NotNil(t, resInstallation)
	assert.Equal(t, "install", resInstallation.RecordKey())
	checkLabelCount(t, collector.LabelCount{"TEST1": 1}, resInstallation.AuthConfig)
	checkLabelCount(t, collector.LabelCount{"id1": 1, "id2": 1}, resInstallation.Users)
	checkLabelCount(t, collector.LabelCount{"NODE_DRIVER_TEST": 1}, resInstallation.NodeDrivers)
	checkLabelCount(t, collector.LabelCount{"KONTAINER_DRIVER_TEST": 1}, resInstallation.KontainerDrivers)
	assert.True(t, resInstallation.HasInternal)
}

func TestInstallationFailListAllForAll(t *testing.T) {
	settings := map[string]string{
		"server-version":     "SERVER_VERSION",
		"telemetry-uid":      "TELEMETRY_UID",
		"ui-default-landing": "UI-DEFAULT-LANDING",
	}
	client := NewInstallationClientMock("FAIL_LIST_ALL",
		"FAIL_LIST_ALL",
		"FAIL_LIST_ALL",
		"FAIL_LIST_ALL",
		"FAIL_LIST_ALL",
		settings)

	c := collector.Installation{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resInstallation := res.(collector.Installation)
	assert.NotNil(t, resInstallation)
	assert.Equal(t, "install", resInstallation.RecordKey())
	checkLabelCount(t, collector.LabelCount{}, resInstallation.AuthConfig)
	checkLabelCount(t, collector.LabelCount{}, resInstallation.Users)
	checkLabelCount(t, collector.LabelCount{}, resInstallation.NodeDrivers)
	checkLabelCount(t, collector.LabelCount{}, resInstallation.KontainerDrivers)
	assert.False(t, resInstallation.HasInternal)
}

func TestInstallationFailSettingByID(t *testing.T) {
	settings := map[string]string{
		"server-version":     "SERVER_VERSION",
		"telemetry-uid":      "TELEMETRY_UID",
		"ui-default-landing": "UI-DEFAULT-LANDING",
	}
	client := NewInstallationClientMock(`
	[
		{
			"id" : "authConfig1",
			"enabled": true,
			"type": "TEST1Config"
		}
	]`,
		`[
		{
			"id" : "user1",
			"type": "user",
			"principalIds" : ["id1://test", "id2://test"]
		}
	]`,
		`[
		{
			"id" : "nodeDriver1",
			"type" : "nodeDriver",
			"active" : true,
			"name" : "NODE_DRIVER_TEST"
		}
	]`,
		`[
		{
			"id" : "kontainerDriver1",
			"type" : "kontainerDriver",
			"active" : true,
			"name" : "KONTAINER_DRIVER_TEST"
		}
	]`,
		`[
			{
				"id" : "cluster1",
				"type": "cluster",
				"internal": true
			}
	]`,
		settings)

	client.Setting = NewSettingOperationsMock("FAIL_BY_ID", settings)
	c := collector.Installation{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resInstallation := res.(collector.Installation)
	assert.NotNil(t, resInstallation)
	assert.Equal(t, "install", resInstallation.RecordKey())
	checkLabelCount(t, collector.LabelCount{"TEST1": 1}, resInstallation.AuthConfig)
	checkLabelCount(t, collector.LabelCount{"id1": 1, "id2": 1}, resInstallation.Users)
	checkLabelCount(t, collector.LabelCount{"NODE_DRIVER_TEST": 1}, resInstallation.NodeDrivers)
	checkLabelCount(t, collector.LabelCount{"KONTAINER_DRIVER_TEST": 1}, resInstallation.KontainerDrivers)
	assert.True(t, resInstallation.HasInternal)
}

func TestInstallationSettingUidEmpty(t *testing.T) {
	settings := map[string]string{
		"server-version":     "SERVER_VERSION",
		"telemetry-uid":      "",
		"ui-default-landing": "UI-DEFAULT-LANDING",
	}
	client := NewInstallationClientMock(`
	[
		{
			"id" : "authConfig1",
			"enabled": true,
			"type": "TEST1Config"
		}
	]`,
		`[
		{
			"id" : "user1",
			"type": "user",
			"principalIds" : ["id1://test", "id2://test"]
		}
	]`,
		`[
		{
			"id" : "nodeDriver1",
			"type" : "nodeDriver",
			"active" : true,
			"name" : "NODE_DRIVER_TEST"
		}
	]`,
		`[
		{
			"id" : "kontainerDriver1",
			"type" : "kontainerDriver",
			"active" : true,
			"name" : "KONTAINER_DRIVER_TEST"
		}
	]`,
		`[
			{
				"id" : "cluster1",
				"type": "cluster",
				"internal": true
			}
	]`,
		settings)

	c := collector.Installation{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resInstallation := res.(collector.Installation)
	assert.NotNil(t, resInstallation)
	assert.Equal(t, "install", resInstallation.RecordKey())
	checkLabelCount(t, collector.LabelCount{"TEST1": 1}, resInstallation.AuthConfig)
	checkLabelCount(t, collector.LabelCount{"id1": 1, "id2": 1}, resInstallation.Users)
	checkLabelCount(t, collector.LabelCount{"NODE_DRIVER_TEST": 1}, resInstallation.NodeDrivers)
	checkLabelCount(t, collector.LabelCount{"KONTAINER_DRIVER_TEST": 1}, resInstallation.KontainerDrivers)
	assert.True(t, resInstallation.HasInternal)
}

func TestInstallationSettingUidEmptyFailSetSetting(t *testing.T) {
	settings := map[string]string{
		"server-version":     "SERVER_VERSION",
		"telemetry-uid":      "",
		"ui-default-landing": "UI-DEFAULT-LANDING",
	}
	client := NewInstallationClientMock(`
	[
		{
			"id" : "authConfig1",
			"enabled": true,
			"type": "TEST1Config"
		}
	]`,
		`[
		{
			"id" : "user1",
			"type": "user",
			"principalIds" : ["id1://test", "id2://test"]
		}
	]`,
		`[
		{
			"id" : "nodeDriver1",
			"type" : "nodeDriver",
			"active" : true,
			"name" : "NODE_DRIVER_TEST"
		}
	]`,
		`[
		{
			"id" : "kontainerDriver1",
			"type" : "kontainerDriver",
			"active" : true,
			"name" : "KONTAINER_DRIVER_TEST"
		}
	]`,
		`[
			{
				"id" : "cluster1",
				"type": "cluster",
				"internal": true
			}
	]`,
		settings)

	client.Setting = NewSettingOperationsMock("FAIL_UPDATE", settings)
	c := collector.Installation{}
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resInstallation := res.(collector.Installation)
	assert.NotNil(t, resInstallation)
	assert.Equal(t, "install", resInstallation.RecordKey())
	checkLabelCount(t, collector.LabelCount{"TEST1": 1}, resInstallation.AuthConfig)
	checkLabelCount(t, collector.LabelCount{"id1": 1, "id2": 1}, resInstallation.Users)
	checkLabelCount(t, collector.LabelCount{"NODE_DRIVER_TEST": 1}, resInstallation.NodeDrivers)
	checkLabelCount(t, collector.LabelCount{"KONTAINER_DRIVER_TEST": 1}, resInstallation.KontainerDrivers)
	assert.True(t, resInstallation.HasInternal)
}

package collector_test

import (
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func TestSettingOK(t *testing.T) {
	client := &rancher.Client{}
	client.Setting = NewSettingOperationsMock(`
		{
			"id": "setting1",
			"type": "setting",
			"default": "DEFAULT_VALUE_TEST",
			"value": "VALUE_TEST"
		}`, nil)
	res := collector.SetSetting(client, "test", "test")
	assert.Nil(t, res)
}

func TestSettingNotFound(t *testing.T) {
	client := &rancher.Client{}
	client.Setting = NewSettingOperationsMock("FAIL_BY_ID_NOT_FOUND", nil)
	res := collector.SetSetting(client, "test", "test")
	assert.Nil(t, res)
}

func TestSettingGetSettingError(t *testing.T) {
	client := &rancher.Client{}
	client.Setting = NewSettingOperationsMock("FAIL_BY_ID", nil)
	res := collector.SetSetting(client, "test", "test")
	assert.NotNil(t, res)
	assert.Equal(t, "[ERROR] SettingOperationsMock ByID Fail", res.Error())
}

func TestSettingCreateError(t *testing.T) {
	client := &rancher.Client{}
	client.Setting = NewSettingOperationsMock("FAIL_BY_ID_NOT_FOUND_CREATE", nil)
	res := collector.SetSetting(client, "test", "test")
	assert.NotNil(t, res)
	assert.Equal(t, "[ERROR] SettingOperationsMock Create Fail", res.Error())
}

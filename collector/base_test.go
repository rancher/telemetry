package collector_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/rancher/norman/clientbase"
	rancherCluster "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/rancher/telemetry/record"
	"github.com/stretchr/testify/assert"
)

type CollectorMock struct {
	Collected bool
	Opts      *collector.CollectorOpts
	Key       string
}

func (c *CollectorMock) RecordKey() string {
	return c.Key
}

func (c *CollectorMock) Collect(opt *collector.CollectorOpts) interface{} {
	c.Opts = opt
	c.Collected = true
	return "Collected" + c.Key
}

func TestBaseGetClusterClient(t *testing.T) {
	prevFunc := collector.NewRancherClusterClient
	defer func() { collector.NewRancherClusterClient = prevFunc }()
	collector.NewRancherClusterClient = func(opts *clientbase.ClientOpts) (*rancherCluster.Client, error) {
		client := &rancherCluster.Client{}
		client.Opts = opts
		return client, nil
	}
	baseClient := &rancher.Client{}
	baseClient.Opts = &clientbase.ClientOpts{}
	testUrlString := fmt.Sprintf("TEST_URL_%d", rand.Int())
	baseClient.Opts.URL = testUrlString
	collectorOpts := &collector.CollectorOpts{baseClient}

	testID := fmt.Sprintf("ID_%d", rand.Int())
	clusterClient, err := collector.GetClusterClient(collectorOpts, testID)
	assert.Nil(t, err)
	assert.NotNil(t, clusterClient)
	clusterClientUrl := fmt.Sprintf("%s/clusters/%s", testUrlString, testID)
	assert.Equal(t, clusterClientUrl, clusterClient.Opts.URL)

	// now if we call GetClusterClient it should return the same URL
	// becase the client is already inserted in the internal map
	clusterClient2, err2 := collector.GetClusterClient(collectorOpts, testID)
	assert.Nil(t, err2)
	assert.NotNil(t, clusterClient2)
	assert.Equal(t, clusterClientUrl, clusterClient2.Opts.URL)
}

func TestBaseGetClusterClientReturnsError(t *testing.T) {
	prevFunc := collector.NewRancherClusterClient
	defer func() { collector.NewRancherClusterClient = prevFunc }()
	collector.NewRancherClusterClient = func(opts *clientbase.ClientOpts) (*rancherCluster.Client, error) {
		return nil, fmt.Errorf("ERROR: I'm a test")
	}
	baseClient := &rancher.Client{}
	baseClient.Opts = &clientbase.ClientOpts{}
	testUrlString := fmt.Sprintf("TEST_URL_%d", rand.Int())
	baseClient.Opts.URL = testUrlString
	collectorOpts := &collector.CollectorOpts{baseClient}

	testID := fmt.Sprintf("ID_%d", rand.Int())
	clusterClient, err := collector.GetClusterClient(collectorOpts, testID)
	assert.NotNil(t, err)
	assert.Nil(t, clusterClient)
	assert.Equal(t, "ERROR: I'm a test", err.Error())
}

func TestBaseGetProjectClient(t *testing.T) {
	prevFunc := collector.NewRancherProjectClient
	defer func() { collector.NewRancherProjectClient = prevFunc }()
	collector.NewRancherProjectClient = func(opts *clientbase.ClientOpts) (*rancherProject.Client, error) {
		client := &rancherProject.Client{}
		client.Opts = opts
		return client, nil
	}
	baseClient := &rancher.Client{}
	baseClient.Opts = &clientbase.ClientOpts{}
	testUrlString := fmt.Sprintf("TEST_URL_%d", rand.Int())
	baseClient.Opts.URL = testUrlString
	collectorOpts := &collector.CollectorOpts{baseClient}

	testID := fmt.Sprintf("ID_%d", rand.Int())
	projectClient, err := collector.GetProjectClient(collectorOpts, testID)
	assert.Nil(t, err)
	assert.NotNil(t, projectClient)
	clusterClientUrl := fmt.Sprintf("%s/projects/%s", testUrlString, testID)
	assert.Equal(t, clusterClientUrl, projectClient.Opts.URL)

	// now if we call GetProjectClient it should return the same URL
	// becase the client is already inserted in the internal map
	projectClient2, err2 := collector.GetProjectClient(collectorOpts, testID)
	assert.Nil(t, err2)
	assert.NotNil(t, projectClient2)
	assert.Equal(t, clusterClientUrl, projectClient2.Opts.URL)
}

func TestBaseGetProjectClientReturnsError(t *testing.T) {
	prevFunc := collector.NewRancherProjectClient
	defer func() { collector.NewRancherProjectClient = prevFunc }()
	collector.NewRancherProjectClient = func(opts *clientbase.ClientOpts) (*rancherProject.Client, error) {
		client := &rancherProject.Client{}
		client.Opts = opts
		return nil, fmt.Errorf("ERROR: I'm a test")
	}
	baseClient := &rancher.Client{}
	baseClient.Opts = &clientbase.ClientOpts{}
	testUrlString := fmt.Sprintf("TEST_URL_%d", rand.Int())
	baseClient.Opts.URL = testUrlString
	collectorOpts := &collector.CollectorOpts{baseClient}

	testID := fmt.Sprintf("ID_%d", rand.Int())
	projectClient, err := collector.GetProjectClient(collectorOpts, testID)
	assert.NotNil(t, err)
	assert.Nil(t, projectClient)
	assert.Equal(t, "ERROR: I'm a test", err.Error())
}

func TestBaseCollectAll(t *testing.T) {
	collector1 := CollectorMock{}
	collector1.Key = "Collector1"
	collector2 := CollectorMock{}
	collector2.Key = "Collector2"
	collector3 := CollectorMock{}
	collector3.Key = "Collector3"

	baseClient := &rancher.Client{}
	baseClient.Opts = &clientbase.ClientOpts{}
	testUrlString := fmt.Sprintf("TEST_URL_%d", rand.Int())
	baseClient.Opts.URL = testUrlString
	baseClient.Catalog = NewCatalogOperationsMock("FAIL_BY_ID_NOT_FOUND")
	baseClient.Project = NewProjectOperationsMock(`[]`)
	baseClient.Cluster = NewClusterOperationsMock(`[]`)
	baseClient.ClusterLogging = NewClusterLoggingOperationsMock(`[]`)
	baseClient.ClusterTemplate = NewClusterTemplateOperationsMock(`[]`)
	baseClient.ClusterTemplateRevision = NewClusterTemplateRevisionOperationsMock(`[]`)
	baseClient.Setting = NewSettingOperationsMock("FAIL_BY_ID", nil)
	baseClient.AuthConfig = NewAuthConfigOperationsMock(`[]`)
	baseClient.User = NewUserOperationsMock(`[]`)
	baseClient.NodeDriver = NewNodeDriverOperationsMock(`[]`)
	baseClient.KontainerDriver = NewKontainerDriverOperationsMock(`[]`)
	baseClient.MultiClusterApp = NewMultiClusterAppOperationsMock(`[]`)
	baseClient.GlobalDnsProvider = NewGlobalDnsProviderOperationsMock(`[]`)
	baseClient.GlobalDns = NewGlobalDnsOperationsMock(`[]`)
	baseClient.Node = NewNodeOperationsMock(`[]`)
	collectorOpts := &collector.CollectorOpts{baseClient}
	record := &record.Record{}
	collector.Register(&collector1)
	collector.Register(&collector2)
	collector.Register(&collector3)

	collector.Run(record, collectorOpts)
	// check that all collectors were called
	assert.True(t, collector1.Collected)
	assert.True(t, collector2.Collected)
	assert.True(t, collector3.Collected)
	colectorResult1, ok := (*record)[collector1.Key]
	assert.True(t, ok)
	assert.Equal(t, "Collected"+collector1.Key, colectorResult1)
	colectorResult2, ok2 := (*record)[collector2.Key]
	assert.True(t, ok2)
	assert.Equal(t, "Collected"+collector2.Key, colectorResult2)
	colectorResult3, ok3 := (*record)[collector3.Key]
	assert.True(t, ok3)
	assert.Equal(t, "Collected"+collector3.Key, colectorResult3)

}

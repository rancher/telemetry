package collector_test

import (
	"encoding/json"
	"fmt"
	"testing"

	rancherCluster "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewClientForTestingCluster(data string) (*rancher.Client, error) {
	client := &rancher.Client{}
	client.Cluster = NewClusterOperationsMock(data)
	client.ClusterLogging = &ClusterLoggingOperationsMock{}
	client.Opts = NewTestClientOpts()
	return client, nil
}

func checkClusterDrivers(t *testing.T, expected collector.LabelCount, actual collector.LabelCount) {
	checkLabelCount(t, expected, actual)
}

func RunClusterTestCase(t *testing.T, clustersJson string,
	namespaces map[string]string,
	expectedTotal int, expectedActive int,
	expectedDrivers collector.LabelCount,
	expectedCPUInfo collector.CpuInfo,
	expectedMemInfo collector.MemoryInfo,
	expectedPodInfo collector.PodInfo,
	expectedNamespaceInfo collector.NsInfo) collector.Cluster {

	var clusters []rancher.Cluster
	err := json.Unmarshal([]byte(clustersJson), &clusters)
	assert.Nil(t, err)
	// set namespaces for each cluster defined
	for _, cluster := range clusters {
		namespace, ok := namespaces[cluster.ID]
		assert.True(t, ok)
		namespaceClient, _ := NewClientForTestingNamespace(namespace)
		collector.ClusterClients[cluster.ID] = namespaceClient
	}
	c := collector.Cluster{}
	client, _ := NewClientForTestingCluster(clustersJson)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)
	// clusters totals and active
	assert.Equal(t, expectedTotal, resCluster.Total)
	assert.Equal(t, expectedActive, resCluster.Active)

	checkClusterDrivers(t, expectedDrivers, resCluster.Driver)
	checkCPUInfo(t, expectedCPUInfo, *resCluster.Cpu)
	checkMemoryInfo(t, expectedMemInfo, *resCluster.Mem)
	checkPodInfo(t, expectedPodInfo, *resCluster.Pod)
	checkNsInfo(t, expectedNamespaceInfo, *resCluster.Ns)

	assert.Equal(t, "cluster", resCluster.RecordKey())

	return resCluster
}
func TestBasicOneSimpleCluster(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"state": "active"
			}
		]`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)
	// clusters totals and active
	assert.Equal(t, 1, resCluster.Total)
	assert.Equal(t, 1, resCluster.Active)

	// Driver
	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)

	// check for info parsed CPU
	// 200m from 4 cores (4000m) --> 5% util
	checkCPUInfo(t, collector.CpuInfo{4, 4, 4, 5, 5, 5}, *resCluster.Cpu)

	// check memory
	// memory is rounded.
	// 8040912 Ki would be 7852,453125 Mi
	// util memory is rounded (140Mi from 7852mi --> 1.7829% used)
	checkMemoryInfo(t, collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2}, *resCluster.Mem)

	// pods
	// 14 out of 220 --> 6.36 (6 rounded)
	checkPodInfo(t, collector.PodInfo{220, 220, 220, 6, 6, 6}, *resCluster.Pod)

	// Namespace
	// only counts no project nss
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, *resCluster.Ns)
}

func TestBasic(t *testing.T) {
	clustersJson := `[
		{
			"id": "local",
			"type": "cluster",
			"allocatable": {
				"cpu": "4",
				"memory": "8040912Ki",
				"pods": "220"
			},
			"requested": {
				"cpu": "200m",
				"memory": "140Mi",
				"pods": "14"
			},
			"driver": "k3s",
			"state": "active"
		}
	]`

	namespaces := make(map[string]string)
	namespaces["local"] = `[
		{"id": "namespace1",
		 "type": "namespace"
		},
		{"id": "namespace2",
		 "type": "namespace"
		}
	]`
	RunClusterTestCase(t, clustersJson, namespaces, 1, 1,
		collector.LabelCount{"k3s": 1},
		// 200m from 4 cores (4000m) --> 5% util
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		// memory is rounded.
		// 8040912 Ki would be 7852,453125 Mi
		// util memory is rounded (140Mi from 7852mi --> 1.7829% used)
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		// 14 out of 220 --> 6.36 (6 rounded)
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		// only counts no project nss
		collector.NsInfo{2, 2, 2, 2, 0, 2})
}

func Test2ClustersWithAllValues(t *testing.T) {
	clustersJson := `[
		{
			"id": "local",
			"type": "cluster",
			"allocatable": {
				"cpu": "4",
				"memory": "8040912Ki",
				"pods": "220"
			},
			"requested": {
				"cpu": "200m",
				"memory": "140Mi",
				"pods": "14"
			},
			"driver": "k3s",
			"state": "active",
			"rancherKubernetesEngineConfig" : {
				"cloudProvider" : {
					"name": "superCloudProvider"
				}
			},
			"enableClusterMonitoring" : true
		},
		{
			"id": "local2",
			"type": "cluster",
			"allocatable": {
				"cpu": "8",
				"memory": "8040912Ki",
				"pods": "350"
			},
			"requested": {
				"cpu": "4000m",
				"memory": "500Mi",
				"pods": "70"
			},
			"driver": "k3s",
			"state": "active",
			"rancherKubernetesEngineConfig" : {
				"cloudProvider" : {
					"name": "amazingCloudProvider"
				}
			},
			"istioEnabled" : true
		}
	]`

	namespaces := make(map[string]string)
	namespaces["local"] = `[
		{"id": "namespace1",
		 "type": "namespace"
		},
		{"id": "namespace2",
		 "type": "namespace"
		}
	]`
	namespaces["local2"] = `[
		{"id": "namespace1",
		 "type": "namespace"
		}
	]`
	resCluster := RunClusterTestCase(t, clustersJson, namespaces, 2, 2,
		collector.LabelCount{"k3s": 2},
		collector.CpuInfo{4, 8, 12, 5, 28, 50},
		collector.MemoryInfo{7852, 7852, 15704, 2, 4, 6},
		collector.PodInfo{220, 350, 570, 6, 13, 20},
		collector.NsInfo{1, 2, 3, 2, 0, 3})

	// check cloud provider
	checkLabelCount(t, collector.LabelCount{"superCloudProvider": 1, "amazingCloudProvider": 1}, resCluster.CloudProvider)

	// 1 cluster has monitoring
	assert.Equal(t, 1, resCluster.MonitoringTotal)

	// 1 cluster has istio
	assert.Equal(t, 1, resCluster.IstioTotal)
}

func TestK3sEmbedded(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "true",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload(`[]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3sBased": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToGetProjectClientError(t *testing.T) {
	prevFunc := collector.ClusterGetProjectClient
	defer func() { collector.ClusterGetProjectClient = prevFunc }()
	collector.ClusterGetProjectClient = func(c *collector.CollectorOpts, id string) (*rancherProject.Client, error) {
		return nil, fmt.Errorf("[ERROR] - GetProjectClient error test")
	}
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "true",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload(`[]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToNoSystemProject(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "false",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload(`[]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToProjectIDEmpty(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients[""] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "true",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload(`[]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToWorkload(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "true",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToWorkloadError(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock(
		`[
			{
				"id": "projectID",
				"labels" : { "authz.management.cattle.io/system-project": "true",
			 				 "test" : "false" }
			}
		]`)
	workloadClientMock, _ := NewClientForTestingWorkload("FAIL_LIST")
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestNoK3sEmbeddedDueToProjectsListError(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"internal" : true,
				"state": "active"
			}
		]`)
	client.Project = NewProjectOperationsMock("FAIL_LIST")
	workloadClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	collector.ProjectClients["projectID"] = workloadClientMock

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{"k3s": 1}, resCluster.Driver)
}

func TestClusterLoggingEmpty(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"state": "active"
			}
		]`)
	client.ClusterLogging = NewClusterLoggingOperationsMock(`[]`)

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	checkClusterDrivers(t, collector.LabelCount{}, resCluster.LogProviderCount)
}

func TestClusterLoggingAllValues(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient

	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"state": "active"
			}
		]`)
	client.ClusterLogging = NewClusterLoggingOperationsMock(
		`[
			{
				"id" : "loggingCustom",
				"appliedSpec" : {
					"customTargetConfig" : {
						"content" : "customTest"
					}
				}
			},
			{
				"id" : "elasticsearchConfig",
				"appliedSpec" : {
					"elasticsearchConfig" : {
						"content" : "elasticsearchConfigTest"
					}
				}
			},
			{
				"id" : "splunkConfig",
				"appliedSpec" : {
					"splunkConfig" : {
						"content" : "splunkConfigTest"
					}
				}
			},
			{
				"id" : "kafkaConfig",
				"appliedSpec" : {
					"kafkaConfig" : {
						"content" : "kafkaConfigTest"
					}
				}
			},
			{
				"id" : "syslogConfig",
				"appliedSpec" : {
					"syslogConfig" : {
						"content" : "syslogConfigTest"
					}
				}
			},
			{
				"id" : "fluentForwarderConfig",
				"appliedSpec" : {
					"fluentForwarderConfig" : {
						"content" : "fluentForwarderConfigTest"
					}
				}
			}
		]`)

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resCluster := res.(collector.Cluster)

	loggingLabels := collector.LabelCount{
		"Elasticsearch": 1,
		"Splunk":        1,
		"Kafka":         1,
		"Syslog":        1,
		"Fluentd":       1,
		"Custom":        1,
	}

	checkClusterDrivers(t, loggingLabels, resCluster.LogProviderCount)
}

func TestClusterSkip(t *testing.T) {
	clustersJson := `[
		{
			"id": "local",
			"type": "cluster",
			"allocatable": {
				"cpu": "0",
				"memory": "0",
				"pods": "0"
			},
			"requested": {
				"cpu": "200m",
				"memory": "140Mi",
				"pods": "14"
			},
			"driver": "k3s",
			"state": "active"
		}
	]`

	namespaces := make(map[string]string)
	namespaces["local"] = `[
		{"id": "namespace1",
		 "type": "namespace"
		}
	]`
	RunClusterTestCase(t, clustersJson, namespaces, 1, 1,
		collector.LabelCount{},
		collector.CpuInfo{0, 0, 0, 0, 0, 0},
		collector.MemoryInfo{0, 0, 0, 0, 0, 0},
		collector.PodInfo{0, 0, 0, 0, 0, 0},
		collector.NsInfo{0, 0, 0, 0, 0, 0})
}

func TestFailClusterListAll(t *testing.T) {
	c := collector.Cluster{}
	client, _ := NewClientForTestingCluster("FAIL_LIST_ALL")
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestLoggingListAll(t *testing.T) {
	c := collector.Cluster{}
	// cluster collect only updates the number of namespaces
	// so we're not interested in the detailed info
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingCluster(
		`[
			{
				"id": "local",
				"type": "cluster",
				"allocatable": {
					"cpu": "4",
					"memory": "8040912Ki",
					"pods": "220"
				},
				"requested": {
					"cpu": "200m",
					"memory": "140Mi",
					"pods": "14"
				},
				"driver": "k3s",
				"state": "active"
			}
		]`)
	client.ClusterLogging = NewClusterLoggingOperationsMock("FAIL_LIST_ALL")
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	// although LoggingListAll failed we keep collecting
	// changed in https://github.com/rancher/telemetry/pull/83
	assert.NotNil(t, res)
}

func TestFailNamespace(t *testing.T) {
	clustersJson := `[
		{
			"id": "local",
			"type": "cluster",
			"name": "TEST_CLUSTER",
			"allocatable": {
				"cpu": "4",
				"memory": "8040912Ki",
				"pods": "220"
			},
			"requested": {
				"cpu": "200m",
				"memory": "140Mi",
				"pods": "14"
			},
			"driver": "k3s",
			"state": "active"
		}
	]`

	namespaces := make(map[string]string)
	namespaces["local"] = "FAIL_LIST_ALL"
	RunClusterTestCase(t, clustersJson, namespaces, 1, 1,
		collector.LabelCount{"k3s": 1},
		// 200m from 4 cores (4000m) --> 5% util
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		// memory is rounded.
		// 8040912 Ki would be 7852,453125 Mi
		// util memory is rounded (140Mi from 7852mi --> 1.7829% used)
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		// 14 out of 220 --> 6.36 (6 rounded)
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		// nothing here (because of namespace error)
		collector.NsInfo{0, 0, 0, 0, 0, 0})
}

func TestFailNamespaceGetClusterClient(t *testing.T) {
	prevFunc := collector.ClusterGetClusterClient
	defer func() { collector.ClusterGetClusterClient = prevFunc }()
	collector.ClusterGetClusterClient = func(c *collector.CollectorOpts, id string) (*rancherCluster.Client, error) {
		return nil, fmt.Errorf("ERROR: GetClusterClient Test")
	}
	clustersJson := `[
		{
			"id": "local",
			"type": "cluster",
			"name": "TEST_CLUSTER",
			"allocatable": {
				"cpu": "4",
				"memory": "8040912Ki",
				"pods": "220"
			},
			"requested": {
				"cpu": "200m",
				"memory": "140Mi",
				"pods": "14"
			},
			"driver": "k3s",
			"state": "active"
		}
	]`

	namespaces := make(map[string]string)
	namespaces["local"] = `[
		{"id": "namespace1",
		 "type": "namespace"
		}
	]`
	RunClusterTestCase(t, clustersJson, namespaces, 1, 1,
		collector.LabelCount{"k3s": 1},
		// 200m from 4 cores (4000m) --> 5% util
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		// memory is rounded.
		// 8040912 Ki would be 7852,453125 Mi
		// util memory is rounded (140Mi from 7852mi --> 1.7829% used)
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		// 14 out of 220 --> 6.36 (6 rounded)
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		// nothing here (because of namespace error)
		collector.NsInfo{0, 0, 0, 0, 0, 0})
}

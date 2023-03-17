package collector_test

import (
	"testing"

	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewClientForTestingNode(data string) (*rancher.Client, error) {
	client := &rancher.Client{}
	client.Node = NewNodeOperationsMock(data)
	return client, nil
}

func RunCollectorNodeTestCase(t *testing.T,
	nodesJson string,
	nodeTemplateJson string,
	expectedActive int,
	expectedImported int,
	expectedCPUInfo collector.CpuInfo,
	expectedMemInfo collector.MemoryInfo,
	expectedPodInfo collector.PodInfo,
	expectedKernels collector.LabelCount,
	expectedKubelets collector.LabelCount,
	expectedKubeproxies collector.LabelCount,
	expectedOss collector.LabelCount,
	expectedDockers collector.LabelCount,
	expectedDrivers collector.LabelCount,
	expectedRoles collector.LabelCount) collector.Node {
	n := collector.Node{}
	client, _ := NewClientForTestingNode(nodesJson)
	client.NodeTemplate = NewNodeTemplateOperationsMock(nodeTemplateJson)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := n.Collect(opts)
	assert.NotNil(t, res)
	resNode := res.(collector.Node)

	checkCPUInfo(t, expectedCPUInfo, resNode.Cpu)
	checkMemoryInfo(t, expectedMemInfo, resNode.Mem)
	checkPodInfo(t, expectedPodInfo, resNode.Pod)
	checkLabelCount(t, expectedKernels, resNode.Kernel)
	checkLabelCount(t, expectedKubelets, resNode.Kubelet)
	checkLabelCount(t, expectedKubeproxies, resNode.Kubeproxy)
	checkLabelCount(t, expectedOss, resNode.Os)
	checkLabelCount(t, expectedDockers, resNode.Docker)
	checkLabelCount(t, expectedDrivers, resNode.Driver)
	checkLabelCount(t, expectedRoles, resNode.Role)
	assert.Equal(t, "node", resNode.RecordKey())
	return resNode
}

func TestNodeBasic(t *testing.T) {
	n := collector.Node{}
	client, _ := NewClientForTestingNode(
		`[
			{
				"id": "local",
				"type": "node",
				"name": "TEST_NODE",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`)
	client.NodeTemplate = NewNodeTemplateOperationsMock(`{
				"id": "nodeTemplate",
			    "driver": "NODE_DRIVER"
			}`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := n.Collect(opts)
	assert.NotNil(t, res)
	resNode := res.(collector.Node)

	assert.Equal(t, 1, resNode.Active)
	assert.Equal(t, 1, resNode.Imported)
	checkCPUInfo(t, collector.CpuInfo{4, 4, 4, 5, 5, 5}, resNode.Cpu)
	checkMemoryInfo(t, collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2}, resNode.Mem)
	checkPodInfo(t, collector.PodInfo{220, 220, 220, 6, 6, 6}, resNode.Pod)
	checkLabelCount(t, collector.LabelCount{"5.3.18-150300.59.49-default": 1}, resNode.Kernel)
	checkLabelCount(t, collector.LabelCount{"v1.23.6+k3s1": 1}, resNode.Kubelet)
	checkLabelCount(t, collector.LabelCount{"v1.23.6+k3s1": 1}, resNode.Kubeproxy)
	checkLabelCount(t, collector.LabelCount{"OpenSuse Leap 15.3": 1}, resNode.Os)
	checkLabelCount(t, collector.LabelCount{"containerd://1.5.11-k3s2": 1}, resNode.Docker)
	checkLabelCount(t, collector.LabelCount{"NODE_DRIVER": 1}, resNode.Driver)
	checkLabelCount(t, collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 1}, resNode.Role)
}

func TestNodeBasicWithTestFunction(t *testing.T) {
	RunCollectorNodeTestCase(
		t,
		`[
			{
				"id": "local",
				"type": "node",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`,
		`{
				"id": "nodeTemplate",
			    "driver": "NODE_DRIVER"
		}`,
		1,
		1,
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		collector.LabelCount{"5.3.18-150300.59.49-default": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"OpenSuse Leap 15.3": 1},
		collector.LabelCount{"containerd://1.5.11-k3s2": 1},
		collector.LabelCount{"NODE_DRIVER": 1},
		collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 1})
}

func TestNode2Nodes(t *testing.T) {
	RunCollectorNodeTestCase(
		t,
		`[
			{
				"id": "local",
				"type": "node",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			},
			{
				"id": "local2",
				"type": "node",
				"allocatable": {
					"cpu": "12",
					"memory": "9640912Ki",
					"pods": "350"
				},
				"requested": {
					"cpu": "4000m",
					"memory": "800Mi",
					"pods": "140"
				},
				"info" : {
					"os": {
						"dockerVersion" : "TEST_DOCKER",
						"kernelVersion" : "TEST_KERNEL",
						"operatingSystem" : "TEST_OS"
					},
					"kubernetes" : {
						"kubeProxyVersion": "TEST_PROXY",
						"kubeletVersion" : "TEST_KUBELET"
					},
					"memory": {
						"memTotalKiB" : 8020488
					},
					"cpu":{
						"count":12
					 }
				},
				"controlPlane": false,
				"etcd": false,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`,
		`{
				"id": "nodeTemplate",
			    "driver": "NODE_DRIVER"
		}`,
		1,
		1,
		collector.CpuInfo{4, 12, 16, 5, 19, 33},
		collector.MemoryInfo{7852, 9414, 17266, 2, 5, 8},
		collector.PodInfo{220, 350, 570, 6, 23, 40},
		collector.LabelCount{"5.3.18-150300.59.49-default": 1, "TEST_KERNEL": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1, "TEST_KUBELET": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1, "TEST_PROXY": 1},
		collector.LabelCount{"OpenSuse Leap 15.3": 1, "TEST_OS": 1},
		collector.LabelCount{"containerd://1.5.11-k3s2": 1, "TEST_DOCKER": 1},
		collector.LabelCount{"NODE_DRIVER": 2},
		collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 2})
}

func TestNode3Nodes1Ignored(t *testing.T) {
	RunCollectorNodeTestCase(
		t,
		`[
			{
				"id": "local",
				"type": "node",
				"hostname": "TEST_HOSTNAME",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			},
			{
				"id": "local2",
				"type": "node",
				"allocatable": {
					"cpu": "12",
					"memory": "9640912Ki",
					"pods": "350"
				},
				"requested": {
					"cpu": "4000m",
					"memory": "800Mi",
					"pods": "140"
				},
				"info" : {
					"os": {
						"dockerVersion" : "TEST_DOCKER",
						"kernelVersion" : "TEST_KERNEL",
						"operatingSystem" : "TEST_OS"
					},
					"kubernetes" : {
						"kubeProxyVersion": "TEST_PROXY",
						"kubeletVersion" : "TEST_KUBELET"
					},
					"memory": {
						"memTotalKiB" : 8020488
					},
					"cpu":{
						"count":12
					 }
				},
				"controlPlane": false,
				"etcd": false,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			},
			{
				"id": "local3",
				"type": "node",
				"allocatable": {
					"cpu": "0",
					"memory": "9640912Ki",
					"pods": "350"
				},
				"requested": {
					"cpu": "4000m",
					"memory": "800Mi",
					"pods": "140"
				},
				"info" : {
					"os": {
						"dockerVersion" : "TEST_DOCKER",
						"kernelVersion" : "TEST_KERNEL",
						"operatingSystem" : "TEST_OS"
					},
					"kubernetes" : {
						"kubeProxyVersion": "TEST_PROXY",
						"kubeletVersion" : "TEST_KUBELET"
					},
					"memory": {
						"memTotalKiB" : 8020488
					},
					"cpu":{
						"count":12
					 }
				},
				"controlPlane": false,
				"etcd": false,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`,
		`{
				"id": "nodeTemplate",
			    "driver": "NODE_DRIVER"
		}`,
		1,
		1,
		collector.CpuInfo{4, 12, 16, 5, 19, 33},
		collector.MemoryInfo{7852, 9414, 17266, 2, 5, 8},
		collector.PodInfo{220, 350, 570, 6, 23, 40},
		collector.LabelCount{"5.3.18-150300.59.49-default": 1, "TEST_KERNEL": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1, "TEST_KUBELET": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1, "TEST_PROXY": 1},
		collector.LabelCount{"OpenSuse Leap 15.3": 1, "TEST_OS": 1},
		collector.LabelCount{"containerd://1.5.11-k3s2": 1, "TEST_DOCKER": 1},
		collector.LabelCount{"NODE_DRIVER": 2},
		collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 2})
}

func TestNodeTestListAll(t *testing.T) {
	n := collector.Node{}
	client, _ := NewClientForTestingNode("FAIL_LIST_ALL")
	client.NodeTemplate = NewNodeTemplateOperationsMock(`{
				"id": "nodeTemplate",
			    "driver": "NODE_DRIVER"
			}`)
	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := n.Collect(opts)
	assert.Nil(t, res)
}

func TestNodeErrorNodeTemplate(t *testing.T) {
	RunCollectorNodeTestCase(
		t,
		`[
			{
				"id": "local",
				"type": "node",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`,
		"FAIL_BY_ID",
		1,
		1,
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		collector.LabelCount{"5.3.18-150300.59.49-default": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"OpenSuse Leap 15.3": 1},
		collector.LabelCount{"containerd://1.5.11-k3s2": 1},
		collector.LabelCount{},
		collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 1})
}

func TestNodeErrorNodeTemplateNotFound(t *testing.T) {
	RunCollectorNodeTestCase(
		t,
		`[
			{
				"id": "local",
				"type": "node",
				"state": "active",
				"imported": true,
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
				"info" : {
					"os": {
						"dockerVersion" : "containerd://1.5.11-k3s2",
						"kernelVersion" : "5.3.18-150300.59.49-default",
						"operatingSystem" : "OpenSuse Leap 15.3"
					},
					"kubernetes" : {
						"kubeProxyVersion": "v1.23.6+k3s1",
						"kubeletVersion" : "v1.23.6+k3s1"
					},
					"memory": {
						"memTotalKiB" : 4020488
					},
					"cpu":{
						"count":2
					 }
				},
				"controlPlane": true,
				"etcd": true,
				"worker": true,
				"nodeTemplateId": "ID_TEMPLATE"
			}
		]`,
		"FAIL_BY_ID_NOT_FOUND",
		1,
		1,
		collector.CpuInfo{4, 4, 4, 5, 5, 5},
		collector.MemoryInfo{7852, 7852, 7852, 2, 2, 2},
		collector.PodInfo{220, 220, 220, 6, 6, 6},
		collector.LabelCount{"5.3.18-150300.59.49-default": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"v1.23.6+k3s1": 1},
		collector.LabelCount{"OpenSuse Leap 15.3": 1},
		collector.LabelCount{"containerd://1.5.11-k3s2": 1},
		collector.LabelCount{},
		collector.LabelCount{"controlplane": 1, "etcd": 1, "worker": 1})
}

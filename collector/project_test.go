package collector_test

import (
	"fmt"
	"testing"

	rancherCluster "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/rancher/telemetry/collector"
	"github.com/stretchr/testify/assert"
)

func NewClientForTestingProject(data string) (*rancher.Client, error) {
	client := &rancher.Client{}
	client.Project = NewProjectOperationsMock(data)
	return client, nil
}

func TestProjectBasic(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectMoreData(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			},
			{"id": "namespace3",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			},
			{
				"id": "local:project_id_2",
				"type": "project"
			}
		]`)

	projectClientMock1, _ := NewClientForTestingWorkload(`[{"id": "workload1"}, {"id": "workload2"}]`)
	projectClientMock1.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock1.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}, {"id": "sourceprov", "type" : "source_test2"}]`)
	projectClientMock1.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock1.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock1.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock1

	projectClientMock2, _ := NewClientForTestingWorkload(`[{"id": "workload1"}, {"id": "workload2"}, {"id": "workload3"}]`)
	projectClientMock2.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}, {"id": "pipeline2"}]`)
	projectClientMock2.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test_3"}, {"id": "sourceprov", "type" : "source_test4"}]`)
	projectClientMock2.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa3"}, {"id": "hpa4"}]`)
	projectClientMock2.Pod = NewPodOperationsMock(`[{"id": "pod1"}, {"id": "pod2"}, {"id": "pod3"}]`)
	projectClientMock2.App =
		NewAppOperationsMock(`[{"id": "app2", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test2" }]`)

	collector.ProjectClients["local:project_id_2"] = projectClientMock2

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 2, resProject.Total)
	checkNsInfo(t, collector.NsInfo{3, 3, 6, 3, 0, 6}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{2, 3, 5, 3}, resProject.Workload)
	checkPipelineInfo(t,
		collector.PipelineInfo{1, collector.LabelCount{"source_test": 1, "source_test2": 1, "source_test4": 1, "source_test_3": 1}, 3},
		resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 2, 3, 2}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 3, 4, 2}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1, "template_test2": 1}, resProject.LibraryCharts)
}

func TestProjectFailProjectListAll(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject("FAIL_LIST_ALL")

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	assert.Nil(t, res)
}

func TestProjectNoNsInfoDueToClientClusterError(t *testing.T) {
	prevFunc := collector.ProjectGetClusterClient
	defer func() { collector.ProjectGetClusterClient = prevFunc }()
	collector.ProjectGetClusterClient = func(c *collector.CollectorOpts, id string) (*rancherCluster.Client, error) {
		return nil, fmt.Errorf("ERROR: GetClusterClient Test")
	}
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{0, 0, 0, 0, 0, 0}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectNoNsDueToNsListAll(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace("FAIL_LIST_ALL")
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{0, 0, 0, 0, 0, 0}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectOnlyNsInfoDueToGetProjectClientError(t *testing.T) {
	prevFunc := collector.ProjectGetProjectClient
	defer func() { collector.ProjectGetProjectClient = prevFunc }()
	collector.ProjectGetProjectClient = func(c *collector.CollectorOpts, id string) (*rancherProject.Client, error) {
		return nil, fmt.Errorf("ERROR: GetProjectClient Test")
	}
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{0, 0, 0, 0}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{0, collector.LabelCount{}, 0}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{0, 0, 0, 0}, resProject.HPA)
	checkPodData(t, collector.PodData{0, 0, 0, 0}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{}, resProject.LibraryCharts)
}

func TestProjectBadRancherCatalog(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "BAD_CATALOG_URL",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{}, resProject.LibraryCharts)
}

func TestProjectNoWordloadDueToWorkloadListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload("FAIL_LIST_ALL")
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{0, 0, 0, 0}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectNoPipelineDueToListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock("FAIL_LIST_ALL")
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 0}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectSourceCodeProviderListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock("FAIL_LIST_ALL")
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{0, collector.LabelCount{}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectHPAListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock("FAIL_LIST_ALL")
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{0, 0, 0, 0}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectPodListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock("FAIL_LIST_ALL")
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "app://test_app?catalog=ns_test/catalog_test&type=type_test&template=template_test" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{0, 0, 0, 0}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{"template_test": 1}, resProject.LibraryCharts)
}

func TestProjectAppListAllError(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock("FAIL_LIST_ALL")

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{}, resProject.LibraryCharts)
}

func TestProjectBadAppURL(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "BAD_PARSE_URL" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{}, resProject.LibraryCharts)
}

func TestProjectBadAppURL2(t *testing.T) {
	c := collector.Project{}
	namespaceClient, _ := NewClientForTestingNamespace(
		`[ {"id": "namespace1",
		    "type": "namespace"
			},
			{"id": "namespace2",
		    "type": "namespace"
			}
		]`)
	collector.ClusterClients["local"] = namespaceClient
	client, _ := NewClientForTestingProject(
		`[
			{
				"id": "local:project_id",
				"type": "project"
			}
		]`)

	projectClientMock, _ := NewClientForTestingWorkload(`[{"id": "workload"}]`)
	projectClientMock.Pipeline = NewPipelineOperationsMock(`[{"id": "pipeline"}]`)
	projectClientMock.SourceCodeProvider =
		NewSourceCodeProviderOperationsMock(`[{"id": "sourceprov", "type" : "source_test"}]`)
	projectClientMock.HorizontalPodAutoscaler =
		NewHorizontalPodAutoscalerOperationsMock(`[{"id": "hpa"}]`)
	projectClientMock.Pod = NewPodOperationsMock(`[{"id": "pod"}]`)
	projectClientMock.App =
		NewAppOperationsMock(`[{"id": "app", "externalId" : "postgres://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require" }]`)

	collector.ProjectClients["local:project_id"] = projectClientMock

	client.Catalog = NewCatalogOperationsMock(
		`{
			"id": "catalog",
			"url": "https://git.rancher.io/charts",
			"name": "catalog_test"
		}`)
	client.Opts = NewTestClientOpts()

	opts := &collector.CollectorOpts{
		Client: client,
	}
	res := c.Collect(opts)
	resProject := res.(collector.Project)

	assert.Equal(t, "project", resProject.RecordKey())
	assert.Equal(t, 1, resProject.Total)
	checkNsInfo(t, collector.NsInfo{2, 2, 2, 2, 0, 2}, resProject.Ns)
	checkWorkloadInfo(t, collector.WorkloadInfo{1, 1, 1, 1}, resProject.Workload)
	checkPipelineInfo(t, collector.PipelineInfo{1, collector.LabelCount{"source_test": 1}, 1}, resProject.Pipeline)
	checkHPAInfo(t, collector.HPAInfo{1, 1, 1, 1}, resProject.HPA)
	checkPodData(t, collector.PodData{1, 1, 1, 1}, resProject.Pod)
	checkLabelCount(t, collector.LabelCount{}, resProject.LibraryCharts)
}

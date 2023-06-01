package collector_test

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	rancherCluster "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	rancherProject "github.com/rancher/rancher/pkg/client/generated/project/v3"
)

func NewClientForTestingNamespace(data string) (*rancherCluster.Client, error) {
	client := &rancherCluster.Client{}
	client.Namespace = NewNamespaceOperationsMock(data)
	return client, nil
}

func NewClientForTestingWorkload(data string) (*rancherProject.Client, error) {
	client := &rancherProject.Client{}
	client.Workload = NewWorkloadOperationsMock(data)
	return client, nil
}

func NewTestClientOpts() *clientbase.ClientOpts {
	ret := &clientbase.ClientOpts{}
	ret.URL = "test"
	ret.AccessKey = "test"
	ret.SecretKey = "test"
	ret.TokenKey = "test"
	return ret
}

type OperationsMock struct {
	ReturnData []interface{}
}

func (m *OperationsMock) SetData(data string) error {
	err := json.Unmarshal([]byte(data), &m.ReturnData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

type ClusterOperationsMock struct {
	rancher.ClusterOperations
	OperationsMock
	ReturnData  []rancher.Cluster
	FailListAll bool
}

func NewClusterOperationsMock(data string) *ClusterOperationsMock {
	ret := &ClusterOperationsMock{}
	if data == "FAIL_LIST_ALL" {
		ret.FailListAll = true
	} else {
		err := ret.SetData(data)
		if err != nil {
			return nil
		}
	}
	return ret
}

func (c *ClusterOperationsMock) ListAll(opts *types.ListOpts) (*rancher.ClusterCollection, error) {
	if c.FailListAll {
		return nil, fmt.Errorf("[ERROR] ClusterOperationsMock ListAll Fail")
	}
	ret := &rancher.ClusterCollection{}
	ret.Data = c.ReturnData
	return ret, nil
}

func (c *ClusterOperationsMock) SetData(data string) error {
	err := json.Unmarshal([]byte(data), &c.ReturnData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

type ClusterLoggingOperationsMock struct {
	rancher.ClusterLoggingOperations
	ReturnData  []rancher.ClusterLogging
	FailListAll bool
}

func (c *ClusterLoggingOperationsMock) SetData(data string) error {
	err := json.Unmarshal([]byte(data), &c.ReturnData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *ClusterLoggingOperationsMock) ListAll(opts *types.ListOpts) (*rancher.ClusterLoggingCollection, error) {
	if c.FailListAll {
		return nil, fmt.Errorf("[ERROR] ClusterLoggingOperationsMock ListAll Fail")
	}
	ret := &rancher.ClusterLoggingCollection{}
	ret.Data = c.ReturnData
	return ret, nil
}

func NewClusterLoggingOperationsMock(data string) *ClusterLoggingOperationsMock {
	ret := &ClusterLoggingOperationsMock{}
	if data == "FAIL_LIST_ALL" {
		ret.FailListAll = true
	} else {
		err := ret.SetData(data)
		if err != nil {
			return nil
		}
	}
	return ret
}

type NamespaceOperationsMock struct {
	rancherCluster.NamespaceOperations
	ReturnData  []rancherCluster.Namespace
	FailListAll bool
}

func (n *NamespaceOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		n.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &n.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (n *NamespaceOperationsMock) ListAll(opts *types.ListOpts) (*rancherCluster.NamespaceCollection, error) {
	if n.FailListAll {
		return nil, fmt.Errorf("[ERROR] NamespaceOperationsMock ListAll Fail")
	}
	ret := &rancherCluster.NamespaceCollection{}
	ret.Data = n.ReturnData
	return ret, nil
}

func NewNamespaceOperationsMock(data string) *NamespaceOperationsMock {
	ret := &NamespaceOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type ProjectOperationsMock struct {
	rancher.ProjectOperations
	ReturnData  []rancher.Project
	FailList    bool
	FailListAll bool
}

func (p *ProjectOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST" {
		p.FailList = true
	} else if data == "FAIL_LIST_ALL" {
		p.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &p.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (p *ProjectOperationsMock) List(opts *types.ListOpts) (*rancher.ProjectCollection, error) {
	if p.FailList {
		return nil, fmt.Errorf("[ERROR] ProjectOperationsMock List Fail")
	}
	ret := &rancher.ProjectCollection{}
	ret.Data = p.ReturnData
	return ret, nil
}

func (p *ProjectOperationsMock) ListAll(opts *types.ListOpts) (*rancher.ProjectCollection, error) {
	if p.FailListAll {
		return nil, fmt.Errorf("[ERROR] ProjectOperationsMock ListAll Fail")
	}
	ret := &rancher.ProjectCollection{}
	ret.Data = p.ReturnData
	return ret, nil
}

func NewProjectOperationsMock(data string) *ProjectOperationsMock {
	ret := &ProjectOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type WorkloadOperationsMock struct {
	rancherProject.WorkloadOperations
	ReturnData  []rancherProject.Workload
	FailList    bool
	FailListAll bool
}

func (w *WorkloadOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST" {
		w.FailList = true
	} else if data == "FAIL_LIST_ALL" {
		w.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &w.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (w *WorkloadOperationsMock) List(opts *types.ListOpts) (*rancherProject.WorkloadCollection, error) {
	if w.FailList {
		return nil, fmt.Errorf("[ERROR] WorkloadOperationsMock List Fail")
	}
	ret := &rancherProject.WorkloadCollection{}
	ret.Data = w.ReturnData
	return ret, nil
}

func (w *WorkloadOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.WorkloadCollection, error) {
	if w.FailListAll {
		return nil, fmt.Errorf("[ERROR] WorkloadOperationsMock ListAll Fail")
	}
	ret := &rancherProject.WorkloadCollection{}
	ret.Data = w.ReturnData
	return ret, nil
}

func NewWorkloadOperationsMock(data string) *WorkloadOperationsMock {
	ret := &WorkloadOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type NodeOperationsMock struct {
	rancher.NodeOperations
	ReturnData  []rancher.Node
	FailListAll bool
}

func (n *NodeOperationsMock) SetData(data string) error {
	err := json.Unmarshal([]byte(data), &n.ReturnData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (n *NodeOperationsMock) ListAll(opts *types.ListOpts) (*rancher.NodeCollection, error) {
	if n.FailListAll {
		return nil, fmt.Errorf("[ERROR] NodeOperationsMock ListAll Fail")
	}
	ret := &rancher.NodeCollection{}
	ret.Data = n.ReturnData
	return ret, nil
}

func NewNodeOperationsMock(data string) *NodeOperationsMock {
	ret := &NodeOperationsMock{}
	if data == "FAIL_LIST_ALL" {
		ret.FailListAll = true
	} else {
		err := ret.SetData(data)
		if err != nil {
			return nil
		}
	}
	return ret
}

type NodeTemplateOperationsMock struct {
	rancher.NodeTemplateOperations
	ReturnData       rancher.NodeTemplate
	FailByID         bool
	FailByIDNotFound bool
}

func (n *NodeTemplateOperationsMock) SetData(data string) error {
	err := json.Unmarshal([]byte(data), &n.ReturnData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (n *NodeTemplateOperationsMock) ByID(id string) (*rancher.NodeTemplate, error) {
	if n.FailByID {
		return nil, fmt.Errorf("[ERROR] NodeTemplateOperationsMock ByID Fail")
	}
	if n.FailByIDNotFound {
		retErr := &clientbase.APIError{}
		retErr.StatusCode = 404
		return nil, retErr
	}
	ret := &rancher.NodeTemplate{}
	*ret = n.ReturnData
	return ret, nil
}

func NewNodeTemplateOperationsMock(data string) *NodeTemplateOperationsMock {
	ret := &NodeTemplateOperationsMock{}
	if data == "FAIL_BY_ID" {
		ret.FailByID = true
	} else if data == "FAIL_BY_ID_NOT_FOUND" {
		ret.FailByIDNotFound = true
	} else {
		err := ret.SetData(data)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	return ret
}

type CatalogOperationsMock struct {
	rancher.CatalogOperations
	ReturnData       rancher.Catalog
	FailByID         bool
	FailByIDNotFound bool
}

func (c *CatalogOperationsMock) SetData(data string) error {
	if data == "FAIL_BY_ID" {
		c.FailByID = true
	} else if data == "FAIL_BY_ID_NOT_FOUND" {
		c.FailByIDNotFound = true
	} else {
		err := json.Unmarshal([]byte(data), &c.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (c *CatalogOperationsMock) ByID(id string) (*rancher.Catalog, error) {
	if c.FailByID {
		return nil, fmt.Errorf("[ERROR] CatalogOperationsMock ByID Fail")
	}
	if c.FailByIDNotFound {
		retErr := &clientbase.APIError{}
		retErr.StatusCode = 404
		return nil, retErr
	}
	ret := &rancher.Catalog{}
	*ret = c.ReturnData
	return ret, nil
}

func NewCatalogOperationsMock(data string) *CatalogOperationsMock {
	ret := &CatalogOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return ret
}

type PipelineOperationsMock struct {
	rancherProject.PipelineOperations
	ReturnData  []rancherProject.Pipeline
	FailListAll bool
}

func (p *PipelineOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		p.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &p.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (p *PipelineOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.PipelineCollection, error) {
	if p.FailListAll {
		return nil, fmt.Errorf("[ERROR] PipelineOperationsMock ListAll Fail")
	}
	ret := &rancherProject.PipelineCollection{}
	ret.Data = p.ReturnData
	return ret, nil
}

func NewPipelineOperationsMock(data string) *PipelineOperationsMock {
	ret := &PipelineOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type SourceCodeProviderOperationsMock struct {
	rancherProject.SourceCodeProviderOperations
	ReturnData  []rancherProject.SourceCodeProvider
	FailListAll bool
}

func (s *SourceCodeProviderOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		s.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &s.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (s *SourceCodeProviderOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.SourceCodeProviderCollection, error) {
	if s.FailListAll {
		return nil, fmt.Errorf("[ERROR] SourceCodeProviderOperationsMock ListAll Fail")
	}
	ret := &rancherProject.SourceCodeProviderCollection{}
	ret.Data = s.ReturnData
	return ret, nil
}

func NewSourceCodeProviderOperationsMock(data string) *SourceCodeProviderOperationsMock {
	ret := &SourceCodeProviderOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type HorizontalPodAutoscalerOperationsMock struct {
	rancherProject.HorizontalPodAutoscalerOperations
	ReturnData  []rancherProject.HorizontalPodAutoscaler
	FailListAll bool
}

func (h *HorizontalPodAutoscalerOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		h.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &h.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (h *HorizontalPodAutoscalerOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.HorizontalPodAutoscalerCollection, error) {
	if h.FailListAll {
		return nil, fmt.Errorf("[ERROR] HorizontalPodAutoscalerOperationsMock ListAll Fail")
	}
	ret := &rancherProject.HorizontalPodAutoscalerCollection{}
	ret.Data = h.ReturnData
	return ret, nil
}

func NewHorizontalPodAutoscalerOperationsMock(data string) *HorizontalPodAutoscalerOperationsMock {
	ret := &HorizontalPodAutoscalerOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type PodOperationsMock struct {
	rancherProject.PodOperations
	ReturnData  []rancherProject.Pod
	FailListAll bool
}

func (p *PodOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		p.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &p.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (p *PodOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.PodCollection, error) {
	if p.FailListAll {
		return nil, fmt.Errorf("[ERROR] PodOperationsMock ListAll Fail")
	}
	ret := &rancherProject.PodCollection{}
	ret.Data = p.ReturnData
	return ret, nil
}

func NewPodOperationsMock(data string) *PodOperationsMock {
	ret := &PodOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type AppOperationsMock struct {
	rancherProject.AppOperations
	ReturnData  []rancherProject.App
	FailListAll bool
}

func (a *AppOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		a.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &a.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (a *AppOperationsMock) ListAll(opts *types.ListOpts) (*rancherProject.AppCollection, error) {
	if a.FailListAll {
		return nil, fmt.Errorf("[ERROR] AppOperationsMock ListAll Fail")
	}
	ret := &rancherProject.AppCollection{}
	ret.Data = a.ReturnData
	return ret, nil
}

func NewAppOperationsMock(data string) *AppOperationsMock {
	ret := &AppOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type ClusterTemplateOperationsMock struct {
	rancher.ClusterTemplateOperations
	ReturnData  []rancher.ClusterTemplate
	FailListAll bool
}

func (c *ClusterTemplateOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		c.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &c.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (a *ClusterTemplateOperationsMock) ListAll(opts *types.ListOpts) (*rancher.ClusterTemplateCollection, error) {
	if a.FailListAll {
		return nil, fmt.Errorf("[ERROR] ClusterTemplateOperationsMock ListAll Fail")
	}
	ret := &rancher.ClusterTemplateCollection{}
	ret.Data = a.ReturnData
	return ret, nil
}

func NewClusterTemplateOperationsMock(data string) *ClusterTemplateOperationsMock {
	ret := &ClusterTemplateOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type ClusterTemplateRevisionOperationsMock struct {
	rancher.ClusterTemplateRevisionOperations
	ReturnData  []rancher.ClusterTemplateRevision
	FailListAll bool
}

func (c *ClusterTemplateRevisionOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		c.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &c.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (a *ClusterTemplateRevisionOperationsMock) ListAll(opts *types.ListOpts) (*rancher.ClusterTemplateRevisionCollection, error) {
	if a.FailListAll {
		return nil, fmt.Errorf("[ERROR] ClusterTemplateOperationsMock ListAll Fail")
	}
	ret := &rancher.ClusterTemplateRevisionCollection{}
	ret.Data = a.ReturnData
	return ret, nil
}

func NewClusterTemplateRevisionOperationsMock(data string) *ClusterTemplateRevisionOperationsMock {
	ret := &ClusterTemplateRevisionOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type SettingOperationsMock struct {
	rancher.SettingOperations
	ReturnData       rancher.Setting
	FailByID         bool
	FailByIDNotFound bool
	FailUpdate       bool
	FailCreate       bool
	Values           map[string]string
}

func (s *SettingOperationsMock) SetData(data string) error {
	if data == "FAIL_BY_ID" {
		s.FailByID = true
	} else if data == "FAIL_UPDATE" {
		s.FailUpdate = true
	} else if data == "FAIL_CREATE" {
		s.FailCreate = true
	} else if data == "FAIL_BY_ID_NOT_FOUND" {
		s.FailByIDNotFound = true
	} else if data == "FAIL_BY_ID_NOT_FOUND_CREATE" {
		s.FailByIDNotFound = true
		s.FailCreate = true
	} else {
		err := json.Unmarshal([]byte(data), &s.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (s *SettingOperationsMock) SetSettingsValues(values map[string]string) {
	s.Values = values
}

func (s *SettingOperationsMock) ByID(id string) (*rancher.Setting, error) {
	if s.FailByID {
		return nil, fmt.Errorf("[ERROR] SettingOperationsMock ByID Fail")
	}
	if s.FailByIDNotFound {
		retErr := &clientbase.APIError{}
		retErr.StatusCode = 404
		return nil, retErr
	}
	value, ok := s.Values[id]
	if ok == true {
		s.ReturnData.Value = value
	}
	return &s.ReturnData, nil
}

func (s *SettingOperationsMock) Update(existing *rancher.Setting, updates interface{}) (*rancher.Setting, error) {
	if s.FailUpdate {
		return nil, fmt.Errorf("[ERROR] SettingOperationsMock Update Fail")
	}
	return existing, nil
}

func (s *SettingOperationsMock) Create(opts *rancher.Setting) (*rancher.Setting, error) {
	if s.FailCreate {
		return nil, fmt.Errorf("[ERROR] SettingOperationsMock Create Fail")
	}
	return &rancher.Setting{}, nil
}

func NewSettingOperationsMock(data string, settings map[string]string) *SettingOperationsMock {
	ret := &SettingOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	if settings != nil {
		ret.SetSettingsValues(settings)
	}
	return ret
}

type AuthConfigOperationsMock struct {
	rancher.AuthConfigOperations
	ReturnData  []rancher.AuthConfig
	FailListAll bool
}

func (a *AuthConfigOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		a.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &a.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (a *AuthConfigOperationsMock) ListAll(opts *types.ListOpts) (*rancher.AuthConfigCollection, error) {
	if a.FailListAll {
		return nil, fmt.Errorf("[ERROR] AuthConfigOperationsMock ListAll Fail")
	}
	ret := &rancher.AuthConfigCollection{}
	ret.Data = a.ReturnData
	return ret, nil
}

func NewAuthConfigOperationsMock(data string) *AuthConfigOperationsMock {
	ret := &AuthConfigOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type UserOperationsMock struct {
	rancher.UserOperations
	ReturnData  []rancher.User
	FailListAll bool
}

func (u *UserOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		u.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &u.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (u *UserOperationsMock) ListAll(opts *types.ListOpts) (*rancher.UserCollection, error) {
	if u.FailListAll {
		return nil, fmt.Errorf("[ERROR] UserOperationsMock ListAll Fail")
	}
	ret := &rancher.UserCollection{}
	ret.Data = u.ReturnData
	return ret, nil
}

func NewUserOperationsMock(data string) *UserOperationsMock {
	ret := &UserOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type NodeDriverOperationsMock struct {
	rancher.NodeDriverOperations
	ReturnData  []rancher.NodeDriver
	FailListAll bool
}

func (n *NodeDriverOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		n.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &n.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (n *NodeDriverOperationsMock) ListAll(opts *types.ListOpts) (*rancher.NodeDriverCollection, error) {
	if n.FailListAll {
		return nil, fmt.Errorf("[ERROR] NodeDriverOperationsMock ListAll Fail")
	}
	ret := &rancher.NodeDriverCollection{}
	ret.Data = n.ReturnData
	return ret, nil
}

func NewNodeDriverOperationsMock(data string) *NodeDriverOperationsMock {
	ret := &NodeDriverOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type KontainerDriverOperationsMock struct {
	rancher.KontainerDriverOperations
	ReturnData  []rancher.KontainerDriver
	FailListAll bool
}

func (k *KontainerDriverOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		k.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &k.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (k *KontainerDriverOperationsMock) ListAll(opts *types.ListOpts) (*rancher.KontainerDriverCollection, error) {
	if k.FailListAll {
		return nil, fmt.Errorf("[ERROR] KontainerDriverOperationsMock ListAll Fail")
	}
	ret := &rancher.KontainerDriverCollection{}
	ret.Data = k.ReturnData
	return ret, nil
}

func NewKontainerDriverOperationsMock(data string) *KontainerDriverOperationsMock {
	ret := &KontainerDriverOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type MultiClusterAppOperationsMock struct {
	rancher.MultiClusterAppOperations
	ReturnData  []rancher.MultiClusterApp
	FailListAll bool
}

func (m *MultiClusterAppOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		m.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &m.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (m *MultiClusterAppOperationsMock) ListAll(opts *types.ListOpts) (*rancher.MultiClusterAppCollection, error) {
	if m.FailListAll {
		return nil, fmt.Errorf("[ERROR] MultiClusterAppOperationsMock ListAll Fail")
	}
	ret := &rancher.MultiClusterAppCollection{}
	ret.Data = m.ReturnData
	return ret, nil
}

func NewMultiClusterAppOperationsMock(data string) *MultiClusterAppOperationsMock {
	ret := &MultiClusterAppOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type TemplateVersionOperationsMock struct {
	rancher.TemplateVersionOperations
	ReturnData rancher.TemplateVersion
	FailByID   bool
}

func (t *TemplateVersionOperationsMock) SetData(data string) error {
	if data == "FAIL_BY_ID" {
		t.FailByID = true
	} else {
		err := json.Unmarshal([]byte(data), &t.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (t *TemplateVersionOperationsMock) ByID(id string) (*rancher.TemplateVersion, error) {
	if t.FailByID {
		return nil, fmt.Errorf("[ERROR] TemplateVersionOperationsMock ByID Fail")
	}
	return &t.ReturnData, nil
}

func NewTemplateVersionOperationsMock(data string) *TemplateVersionOperationsMock {
	ret := &TemplateVersionOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type GlobalDnsProviderOperationsMock struct {
	rancher.GlobalDnsProviderOperations
	ReturnData  []rancher.GlobalDnsProvider
	FailListAll bool
}

func (g *GlobalDnsProviderOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		g.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &g.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (m *GlobalDnsProviderOperationsMock) ListAll(opts *types.ListOpts) (*rancher.GlobalDnsProviderCollection, error) {
	if m.FailListAll {
		return nil, fmt.Errorf("[ERROR] GlobalDnsProviderOperationsMock ListAll Fail")
	}
	ret := &rancher.GlobalDnsProviderCollection{}
	ret.Data = m.ReturnData
	return ret, nil
}

func NewGlobalDnsProviderOperationsMock(data string) *GlobalDnsProviderOperationsMock {
	ret := &GlobalDnsProviderOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

type GlobalDnsOperationsMock struct {
	rancher.GlobalDnsOperations
	ReturnData  []rancher.GlobalDns
	FailListAll bool
}

func (g *GlobalDnsOperationsMock) SetData(data string) error {
	if data == "FAIL_LIST_ALL" {
		g.FailListAll = true
	} else {
		err := json.Unmarshal([]byte(data), &g.ReturnData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (m *GlobalDnsOperationsMock) ListAll(opts *types.ListOpts) (*rancher.GlobalDnsCollection, error) {
	if m.FailListAll {
		return nil, fmt.Errorf("[ERROR] GlobalDnsOperationsMock ListAll Fail")
	}
	ret := &rancher.GlobalDnsCollection{}
	ret.Data = m.ReturnData
	return ret, nil
}

func NewGlobalDnsOperationsMock(data string) *GlobalDnsOperationsMock {
	ret := &GlobalDnsOperationsMock{}
	err := ret.SetData(data)
	if err != nil {
		return nil
	}
	return ret
}

// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"time"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"goodrain.com/cloud-adaptor/library/bcode"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(rainbondv1alpha1.AddToScheme(scheme))
}

//Cluster cluster
type Cluster struct {
	Name              string                 `json:"name,omitempty"`
	ClusterID         string                 `json:"cluster_id,omitempty"`
	Created           Time                   `json:"created,omitempty"`
	MasterURL         MasterURL              `json:"master_url,omitempty"`
	State             string                 `json:"state,omitempty"`
	ClusterType       string                 `json:"cluster_type,omitempty"`
	CurrentVersion    string                 `json:"current_version,omitempty"`
	RegionID          string                 `json:"region_id,omitempty"`
	ZoneID            string                 `json:"zone_id,omitempty"`
	VPCID             string                 `json:"vpc_id,omitempty"`
	SecurityGroupID   string                 `json:"security_group_id,omitempty"`
	VSwitchID         string                 `json:"vswitch_id,omitempty"`
	NetworkMode       string                 `json:"network_mode,omitempty"`
	SubnetCIDR        string                 `json:"subnet_cidr,omitempty"`
	PodCIDR           string                 `json:"container_cidr,omitempty"`
	DockerVersion     string                 `json:"docker_version,omitempty"`
	KubernetesVersion string                 `json:"kubernetes_version,omitempty"`
	Size              int                    `json:"size,omitempty"`
	Parameters        map[string]interface{} `json:"parameters,omitempty"`
	RainbondInit      bool                   `json:"rainbond_init,omitempty"`
	CreateLogPath     string                 `json:"create_log_path,omitempty"`
	EIP               []string               `json:"eip,omitempty"`
}

//RunningState running
var RunningState = "running"

//OfflineState offline
var OfflineState = "offline"

//InstallingState offline
var InstallingState = "installing"

//InitState -
var InitState = "initial"

//InstallFailed 安装失败
var InstallFailed = "failed"

//Time time
type Time struct {
	timer time.Time
}

//NewTime new time
func NewTime(timer time.Time) Time {
	return Time{timer: timer}
}

//Time time
func (t *Time) Time() time.Time {
	return t.timer
}

//MarshalJSON -
func (t *Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(time.RFC3339)+2)
	b = append(b, '"')
	b = append(b, []byte(t.timer.Format(time.RFC3339))...)
	b = append(b, '"')
	return b, nil
}

//UnmarshalJSON -
func (t *Time) UnmarshalJSON(in []byte) error {
	inStr := string(in)
	if strings.Contains(inStr, "\"") {
		inStr = strings.Replace(inStr, "\"", "", -1)
	}
	t1, err := time.Parse(time.RFC3339, inStr)
	if err != nil {
		return err
	}
	t.timer = t1
	return nil
}

//MasterURL master url struct
type MasterURL struct {
	APIServerEndpoint         string `json:"api_server_endpoint,omitempty"`
	DashboardEndpoint         string `json:"dashboard_endpoint,omitempty"`
	MiranaEndpoint            string `json:"mirana_endpoint,omitempty"`
	ReverseTunnelEndpoint     string `json:"reverse_tunnel_endpoint,omitempty"`
	IntranetAPIServerEndpoint string `json:"intranet_api_server_endpoint,omitempty"`
}

//MarshalJSON -
func (t *MasterURL) MarshalJSON() ([]byte, error) {
	var info = make(map[string]interface{})
	if t.APIServerEndpoint != "" {
		info["api_server_endpoint"] = t.APIServerEndpoint
	}
	if t.DashboardEndpoint != "" {
		info["dashboard_endpoint"] = t.DashboardEndpoint
	}
	if t.IntranetAPIServerEndpoint != "" {
		info["intranet_api_server_endpoint"] = t.IntranetAPIServerEndpoint
	}
	if t.MiranaEndpoint != "" {
		info["mirana_endpoint"] = t.MiranaEndpoint
	}
	if t.ReverseTunnelEndpoint != "" {
		info["reverse_tunnel_endpoint"] = t.ReverseTunnelEndpoint
	}
	return json.Marshal(info)
}

//UnmarshalJSON -
func (t *MasterURL) UnmarshalJSON(in []byte) error {
	inStr := string(in)
	if len(inStr) < 3 {
		return nil
	}
	jsonStr := inStr[1 : len(inStr)-1]
	jsonStr = strings.Replace(jsonStr, "\\", "", -1)
	var info = make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		return err
	}
	if info["api_server_endpoint"] != nil {
		t.APIServerEndpoint = info["api_server_endpoint"].(string)
	}
	if info["dashboard_endpoint"] != nil {
		t.DashboardEndpoint = info["dashboard_endpoint"].(string)
	}
	if info["intranet_api_server_endpoint"] != nil {
		t.IntranetAPIServerEndpoint = info["intranet_api_server_endpoint"].(string)
	}
	if info["mirana_endpoint"] != nil {
		t.MiranaEndpoint = info["mirana_endpoint"].(string)
	}
	if info["reverse_tunnel_endpoint"] != nil {
		t.ReverseTunnelEndpoint = info["reverse_tunnel_endpoint"].(string)
	}
	return nil
}

//CreateClusterConfig cluster config
type CreateClusterConfig interface {
}

//KubernetesClusterConfig kubernetes cluster commmon config
type KubernetesClusterConfig struct {
	EnterpriseID       string   `json:"eid"`
	AccessKey          string   `json:"access_key"`
	SecretKey          string   `json:"secret_key"`
	Provider           string   `json:"provider_name"`
	ClusterName        string   `json:"name"`
	Region             string   `json:"region,omitempty"`
	ClusterCIDR        string   `json:"clusterCIDR,omitempty"`
	ServiceCIDR        string   `json:"serviceCIDR,omitempty"`
	ClusterType        string   `json:"clusterType,omitempty"`
	VpcID              string   `json:"vpcID,omitempty"`
	VSwitchID          string   `json:"vSwitchID,omitempty"`
	NetworkMode        string   `json:"networkMode,omitempty"`
	WorkerResourceType string   `json:"workerResourceType,omitempty"`
	WorkerNodeNum      int      `json:"workerNum,omitempty"`
	Nodes              NodeList `json:"nodes,omitempty"`
	MasterNodeNum      int      `json:"masterNodeNum,omitempty"`
	ETCDNodeNum        int      `json:"etcdNodeNum,omitempty"`
	InstanceType       string   `json:"instanceType,omitempty"`
	DockerVersion      string   `json:"dockerVersion,omitempty"`
	KubernetesVersion  string   `json:"kubernetesVersion,omitempty"`
}

//NodeList node list
type NodeList []ConfigNode

//Validate validate nodes
func (n NodeList) Validate() error {
	if len(n) == 0 {
		return bcode.ErrClusterNodeEmpty
	}
	var masterNode, etcdNode, workerNode int
	for _, node := range n {
		ip := net.ParseIP(node.IP)
		if ip == nil || ip.IsLoopback() {
			return bcode.ErrClusterNodeIPInvalid
		}
		if node.SSHPort < 0 || node.SSHPort > 65533 {
			return bcode.ErrClusterNodePortInvalid
		}
		if strings.Contains(strings.Join(node.Roles, ","), "controlplane") {
			masterNode++
		}
		if strings.Contains(strings.Join(node.Roles, ","), "worker") {
			workerNode++
		}
		if strings.Contains(strings.Join(node.Roles, ","), "etcd") {
			etcdNode++
		}
	}
	if masterNode == 0 || etcdNode == 0 || workerNode == 0 {
		return bcode.ErrClusterNodeRoleMiss
	}
	if etcdNode%2 == 0 {
		return bcode.ErrETCDNodeNotOddNumer
	}
	return nil
}

//ConfigNode config node
type ConfigNode struct {
	IP               string   `json:"ip"`
	InternalAddress  string   `json:"internalIP,omitempty"`
	SSHUser          string   `json:"sshUser,omitempty"`
	SSHPort          int      `json:"sshPort,omitempty"`
	DockerSocketPath string   `json:"dockerSocketPath,omitempty"`
	Roles            []string `json:"roles,omitempty"`
}

//ClusterType 集群类型
type ClusterType string

//ManagedKubernetes 托管集群
var ManagedKubernetes ClusterType = "ManagedKubernetes"

//AckClusterConfig ack cluster config
type AckClusterConfig struct {
	Name                 string      `json:"name,omitempty"`
	ClusterType          ClusterType `json:"cluster_type,omitempty"`
	DisableRollback      bool        `json:"disable_rollback,omitempty"`
	TimeoutMins          int         `json:"timeout_mins,omitempty"`
	KubernetesVersion    string      `json:"kubernetes_version,omitempty"`
	RegionID             string      `json:"region_id,omitempty"`
	SNATEntry            bool        `json:"snat_entry,omitempty"`
	CloudMonitorFlags    bool        `json:"cloud_monitor_flags,omitempty"`
	EndpointPublicAccess bool        `json:"endpoint_public_access,omitempty"`
	//是否开启集群删除保护，防止通过控制台或api误删除集群
	DeletionProtection bool     `json:"deletion_protection,omitempty"`
	NodeCidrMask       string   `json:"node_cidr_mask,omitempty"`
	ProxyMode          string   `json:"proxy_mode,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Addons             []Addon  `json:"addons,omitempty"`
	OSType             string   `json:"os_type,omitempty"`
	Platform           string   `json:"platform,omitempty"`
	Runtime            Runtime  `json:"runtime,omitempty"`
	//Worker实例规格多实例规格参数
	WorkerInstanceType       []string         `json:"worker_instance_types,omitempty"`
	NumOfNodes               int              `json:"num_of_nodes,omitempty"`
	WorkerSystemDiskCategory string           `json:"worker_system_disk_category,omitempty"`
	WorkerSystemDiskSize     int              `json:"worker_system_disk_size,omitempty"`
	WorkerDataDisks          []WorkerDataDisk `json:"worker_data_disks,omitempty"`
	WorkerInstanceChargeType string           `json:"worker_instance_charge_type,omitempty"`
	VPCID                    string           `json:"vpcid,omitempty"`
	ContainerCIDR            string           `json:"container_cidr,omitempty"`
	ServiceCIDR              string           `json:"service_cidr,omitempty"`
	VSwitchIDs               []string         `json:"vswitch_ids,omitempty"`
	LoginPassword            string           `json:"login_password,omitempty"`
	CPUPolicy                string           `json:"cpu_policy,omitempty"`
}

//Addon 选装addon
type Addon struct {
	Name    string `json:"name,omitempty"`
	Disable bool   `json:"disable,omitempty"`
}

//Runtime container runtime
type Runtime struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

//WorkerDataDisk -
type WorkerDataDisk struct {
	Category  string `json:"category,omitempty"`
	Size      string `json:"size,omitempty"`
	Encrypted string `json:"encrypted,omitempty"`
}

//KubeConfig kube config
type KubeConfig struct {
	Config string `json:"config,omitempty"`
}

//ToKubeConfig Converts to a kube config structure
func (c *KubeConfig) ToKubeConfig() (*rest.Config, error) {
	config, err := clientcmd.Load([]byte(c.Config))
	if err != nil {
		return nil, err
	}
	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
}

//KubeServer kube api
func (c *KubeConfig) KubeServer() (string, error) {
	config, err := clientcmd.Load([]byte(c.Config))
	if err != nil {
		return "", err
	}
	for _, cluster := range config.Clusters {
		return cluster.Server, nil
	}
	return "", nil
}

//Save save kubeconfig
func (c *KubeConfig) Save(configpath string) error {
	pDir := path.Dir(configpath)
	if _, err := os.Stat(pDir); os.IsNotExist(err) {
		os.MkdirAll(pDir, 0755)
	}
	if err := ioutil.WriteFile(configpath, []byte(c.Config), 0644); err != nil {
		return err
	}
	return nil
}

//GetKubeClient get kube client
func (c *KubeConfig) GetKubeClient() (*kubernetes.Clientset, client.Client, error) {
	config, err := c.ToKubeConfig()
	if err != nil {
		return nil, nil, err
	}
	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("NewForConfig failure %+v", err)
	}
	mapper, err := apiutil.NewDynamicRESTMapper(config, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, nil, fmt.Errorf("NewDynamicRESTMapper failure %+v", err)
	}
	runtimeClient, err := client.New(config, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, nil, fmt.Errorf("New kube client failure %+v", err)
	}
	return coreClient, runtimeClient, nil
}

// VPC vpc network
type VPC struct {
	VpcID           string `json:"VpcId" xml:"VpcId"`
	RegionID        string `json:"RegionId" xml:"RegionId"`
	Status          string `json:"Status" xml:"Status"`
	VpcName         string `json:"VpcName" xml:"VpcName"`
	CreationTime    string `json:"CreationTime" xml:"CreationTime"`
	CidrBlock       string `json:"CidrBlock" xml:"CidrBlock"`
	Ipv6CidrBlock   string `json:"Ipv6CidrBlock" xml:"Ipv6CidrBlock"`
	VRouterID       string `json:"VRouterId" xml:"VRouterId"`
	Description     string `json:"Description" xml:"Description"`
	IsDefault       bool   `json:"IsDefault" xml:"IsDefault"`
	NetworkACLNum   string `json:"NetworkAclNum" xml:"NetworkAclNum"`
	ResourceGroupID string `json:"ResourceGroupId" xml:"ResourceGroupId"`
	CenStatus       string `json:"CenStatus" xml:"CenStatus"`
	EnableIpv6      bool   `json:"EnableIpv6" xml:"EnableIpv6"`
	Tags            []Tag  `json:"Tags" xml:"Tags"`
}

//VSwitch -
type VSwitch struct {
	VpcID                   string `json:"VpcId" xml:"VpcId"`
	RegionID                string `json:"RegionId" xml:"RegionId"`
	VSwitchID               string `json:"VSwitchId" xml:"VSwitchId"`
	Status                  string `json:"Status" xml:"Status"`
	CidrBlock               string `json:"CidrBlock" xml:"CidrBlock"`
	Ipv6CidrBlock           string `json:"Ipv6CidrBlock" xml:"Ipv6CidrBlock"`
	ZoneID                  string `json:"ZoneId" xml:"ZoneId"`
	AvailableIPAddressCount int64  `json:"AvailableIpAddressCount" xml:"AvailableIpAddressCount"`
	Description             string `json:"Description" xml:"Description"`
	VSwitchName             string `json:"VSwitchName" xml:"VSwitchName"`
	CreationTime            string `json:"CreationTime" xml:"CreationTime"`
	IsDefault               bool   `json:"IsDefault" xml:"IsDefault"`
	ResourceGroupID         string `json:"ResourceGroupId" xml:"ResourceGroupId"`
	NetworkACLID            string `json:"NetworkAclId" xml:"NetworkAclId"`
	Tags                    []Tag  `json:"Tags" xml:"Tags"`
}

// Zone is a nested struct in vpc response
type Zone struct {
	ZoneID    string `json:"ZoneId" xml:"ZoneId"`
	LocalName string `json:"LocalName" xml:"LocalName"`
}

//Tag TAG
type Tag struct {
	Key   string `json:"Key" xml:"Key"`
	Value string `json:"Value" xml:"Value"`
}

//InstanceType worker instance type
type InstanceType struct {
	InstanceTypeID     string  `json:"InstanceTypeId" xml:"InstanceTypeId"`
	CPUCoreCount       int     `json:"CpuCoreCount" xml:"CpuCoreCount"`
	MemorySize         float64 `json:"MemorySize" xml:"MemorySize"`
	InstanceTypeFamily string  `json:"InstanceTypeFamily" xml:"InstanceTypeFamily"`
}

//Database database
type Database struct {
	Name       string `json:"name,omitempty"`
	RegionID   string `json:"regionID,omitempty"`
	ZoneID     string `json:"zoneID,omitempty"`
	InstanceID string `json:"instanceID,omitempty"`
	UserName   string `json:"userName,omitempty"`
	Password   string `json:"password,omitempty"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	PodCIDR    string `json:"podCIDR,omitempty"`
	VPCID      string `json:"vpcID,omitempty"`
	VSwitchID  string `json:"vSwitchId"`
	ClusterID  string `json:"clusterID"`
}

//RainbondInitConfig rainbond init config
type RainbondInitConfig struct {
	EnableHA          bool
	RainbondVersion   string
	RainbondCIVersion string
	ClusterID         string
	RegionDatabase    *Database
	ETCDConfig        *rainbondv1alpha1.EtcdConfig
	NasServer         string
	SuffixHTTPHost    string
	GatewayNodes      []*rainbondv1alpha1.K8sNode
	ChaosNodes        []*rainbondv1alpha1.K8sNode
	EIPs              []string
}

//NasStorageInfo nas storage info
type NasStorageInfo struct {
	FileSystemID string `json:"FileSystemId" xml:"FileSystemId"`
	Description  string `json:"Description" xml:"Description"`
	CreateTime   string `json:"CreateTime" xml:"CreateTime"`
	RegionID     string `json:"RegionId" xml:"RegionId"`
	ProtocolType string `json:"ProtocolType" xml:"ProtocolType"`
	StorageType  string `json:"StorageType" xml:"StorageType"`
	MeteredSize  int64  `json:"MeteredSize" xml:"MeteredSize"`
	ZoneID       string `json:"ZoneId" xml:"ZoneId"`
	Bandwidth    int64  `json:"Bandwidth" xml:"Bandwidth"`
	Capacity     int64  `json:"Capacity" xml:"Capacity"`
	Status       string `json:"Status" xml:"Status"`
}

// LoadBalancer is a nested struct in slb response
type LoadBalancer struct {
	LoadBalancerID     string `json:"LoadBalancerId" xml:"LoadBalancerId"`
	LoadBalancerName   string `json:"LoadBalancerName" xml:"LoadBalancerName"`
	LoadBalancerStatus string `json:"LoadBalancerStatus" xml:"LoadBalancerStatus"`
	Address            string `json:"Address" xml:"Address"`
	AddressType        string `json:"AddressType" xml:"AddressType"`
	RegionID           string `json:"RegionId" xml:"RegionId"`
	RegionIDAlias      string `json:"RegionIdAlias" xml:"RegionIdAlias"`
	VSwitchID          string `json:"VSwitchId" xml:"VSwitchId"`
	VpcID              string `json:"VpcId" xml:"VpcId"`
	NetworkType        string `json:"NetworkType" xml:"NetworkType"`
	MasterZoneID       string `json:"MasterZoneId" xml:"MasterZoneId"`
	SlaveZoneID        string `json:"SlaveZoneId" xml:"SlaveZoneId"`
	InternetChargeType string `json:"InternetChargeType" xml:"InternetChargeType"`
	CreateTime         string `json:"CreateTime" xml:"CreateTime"`
	CreateTimeStamp    int64  `json:"CreateTimeStamp" xml:"CreateTimeStamp"`
	PayType            string `json:"PayType" xml:"PayType"`
	ResourceGroupID    string `json:"ResourceGroupId" xml:"ResourceGroupId"`
	AddressIPVersion   string `json:"AddressIPVersion" xml:"AddressIPVersion"`
}

//RainbondRegionStatus rainbond region status
type RainbondRegionStatus struct {
	OperatorReady   bool
	RainbondCluster *rainbondv1alpha1.RainbondCluster
	RainbondPackage *rainbondv1alpha1.RainbondPackage
	RainbondVolume  *rainbondv1alpha1.RainbondVolume
	RegionConfig    *v1.ConfigMap
}

//AvailableResourceZone available resource
type AvailableResourceZone struct {
	Status         string `json:"status"`
	StatusCategory string `json:"statusCategory"`
	ZoneID         string `json:"zoneID"`
}

//ExpansionNode expansion node
type ExpansionNode struct {
	EnterpriseID       string   `json:"eid"`
	Provider           string   `json:"provider"`
	AccessKey          string   `json:"accessKey"`
	SecretKey          string   `json:"secretKey"`
	ClusterID          string   `json:"clusterID"`
	Nodes              NodeList `json:"nodes,omitempty"`
	WorkerResourceType string   `json:"workerResourceType,omitempty"`
	WorkerNodeNum      int      `json:"workerNum,omitempty"`
	MasterNodeNum      int      `json:"masterNodeNum,omitempty"`
	ETCDNodeNum        int      `json:"etcdNodeNum,omitempty"`
	InstanceType       string   `json:"instanceType,omitempty"`
	DockerVersion      string   `json:"dockerVersion,omitempty"`
}

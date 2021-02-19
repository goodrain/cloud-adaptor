package models

//RKECluster RKE cluster
type RKECluster struct {
	Model
	Name              string `gorm:"column:name" json:"name,omitempty"`
	ClusterID         string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	APIURL            string `gorm:"column:apiURL,type:text" json:"apiURL,omitempty"`
	KubeConfig        string `gorm:"column:kubeConfig,type:text" json:"kubeConfig,omitempty"`
	NetworkMode       string `gorm:"column:networkMode" json:"networkMode,omitempty"`
	ServiceCIDR       string `gorm:"column:serviceCIDR" json:"serviceCIDR,omitempty"`
	PodCIDR           string `gorm:"column:podCIDR" json:"podCIDR,omitempty"`
	KubernetesVersion string `gorm:"column:kubernetesVersion" json:"kubernetesVersion,omitempty"`
	RainbondInit      bool   `gorm:"column:rainbondInit" json:"rainbondInit,omitempty"`
	CreateLogPath     string `gorm:"column:createLogPath" json:"createLogPath,omitempty"`
	NodeList          string `gorm:"column:nodeList,type:text" json:"nodeList,omitempty"`
	Stats             string `gorm:"column:stats" json:"stats,omitempty"`
}

//CustomCluster custom cluster
type CustomCluster struct {
	Model
	Name       string `gorm:"column:name" json:"name,omitempty"`
	ClusterID  string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	KubeConfig string `gorm:"column:kubeConfig,type:text" json:"kubeConfig,omitempty"`
	EIP        string `gorm:"column:eip" json:"eip,omitempty"`
}

//RainbondClusterConfig rainbond cluster config
type RainbondClusterConfig struct {
	Model
	ClusterID string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	Config    string `gorm:"column:config,type:text" json:"config,omitempty"`
}

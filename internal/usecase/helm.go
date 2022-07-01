package usecase

// ChartInfo -
type ChartInfo struct {
	// 是否开启高可用
	EnableHA bool `json:"enableHA"`
	// 是否使用外部镜像镜像仓库
	ImageHub *ImageHub `json:"imageHub"`
	// 外部ETCD
	Etcd *Etcd `json:"etcd"`
	// 外部存储
	Estorage *Estorage `json:"estorage"`
	// 数据库
	Database *Database `json:"database"`
	// 构建节点
	NodesForChaos *NodesForChaos `json:"nodesForChaos"`
	// 网关节点
	NodesForGateway *NodesForGateway `json:"nodesForGateway"`
	// 网关地址
	GatewayIngressIPs string `json:"gatewayIngressIPs"`
	// 是否安装控制台
	AppUI bool `json:"appui"`
	// token标识
	Token string `json:"token"`
	// 企业id
	EID string `json:"eid"`
	// 控制台域名
	Domain string `json:"domain"`
	// 对接类型
	DockingType string `json:"dockingType"`
	// 云服务
	CloudServer string `json:"cloudserver"`
}

// ImageHub -
type ImageHub struct {
	Enable    bool   `json:"enable"`
	Domain    string `json:"domain"`
	Namespace string `json:"namespace"`
	Password  string `json:"password"`
	Username  string `json:"username"`
}

// Etcd -
type Etcd struct {
	Enable     bool   `json:"enable"`
	Endpoints  []*Eip `json:"endpoints"`
	SecretName string `json:"secretName"`
}

// Eip -
type Eip struct {
	IP string `json:"ip"`
}

// Estorage -
type Estorage struct {
	Enable bool `json:"enable"`
	RWX    *RWX `json:"rwx"`
	RWO    *RWO `json:"rwo"`
	NFS    *NFS `json:"nfs"`
}

// RWX -
type RWX struct {
	Enable bool       `json:"enable"`
	Config *RWXConfig `json:"config"`
}

// RWXConfig -
type RWXConfig struct {
	StorageClassName string `json:"storageClassName"`
	Server           string `json:"server"`
}

// RWO -
type RWO struct {
	Enable           bool   `json:"enable"`
	StorageClassName string `json:"storageClassName"`
}

// Database -
type Database struct {
	Enable         bool            `json:"enable"`
	UIDatabase     *UIDatabase     `json:"uiDatabase"`
	RegionDatabase *RegionDatabase `json:"regionDatabase"`
}

// UIDatabase -
type UIDatabase struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

// RegionDatabase -
type RegionDatabase struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

// NodesForChaos -
type NodesForChaos struct {
	Enable bool         `json:"enable"`
	Nodes  []*ChaosNode `json:"nodes"`
}

// ChaosNode -
type ChaosNode struct {
	Name string `json:"name"`
}

// NodesForGateway -
type NodesForGateway struct {
	Enable bool           `json:"enable"`
	Nodes  []*GatewayNode `json:"nodes"`
}

// GatewayNode -
type GatewayNode struct {
	ExternalIP string `json:"externalIP"`
	InternalIP string `json:"InternalIP"`
	Name       string `json:"name"`
}

// NFS -
type NFS struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

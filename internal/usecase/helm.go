package usecase

// 定义接收数据的结构体

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

type ImageHub struct {
	Enable    bool   `json:"enable"`
	Domain    string `json:"domain"`
	Namespace string `json:"namespace"`
	Password  string `json:"password"`
	Username  string `json:"username"`
}

type Etcd struct {
	Enable     bool   `json:"enable"`
	Endpoints  []*Eip `json:"endpoints"`
	SecretName string `json:"secretName"`
}

type Eip struct {
	Ip string `json:"ip"`
}

type Estorage struct {
	Enable bool `json:"enable"`
	RWX    *RWX `json:"rwx"`
	RWO    *RWO `json:"rwo"`
	NFS    *NFS `json:"nfs"`
}

type RWX struct {
	Enable bool       `json:"enable"`
	Config *RWXConfig `json:"config"`
}

type RWXConfig struct {
	StorageClassName string `json:"storageClassName"`
	Server           string `json:"server"`
}

type RWO struct {
	Enable           bool   `json:"enable"`
	StorageClassName string `json:"storageClassName"`
}

type Database struct {
	Enable         bool            `json:"enable"`
	UiDatabase     *UiDatabase     `json:"uiDatabase"`
	RegionDatabase *RegionDatabase `json:"regionDatabase"`
}

// 控制台数据库

type UiDatabase struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

// 数据中心数据库

type RegionDatabase struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

type NodesForChaos struct {
	Enable bool         `json:"enable"`
	Nodes  []*ChaosNode `json:"nodes"`
}

type ChaosNode struct {
	Name string `json:"name"`
}

type NodesForGateway struct {
	Enable bool           `json:"enable"`
	Nodes  []*GatewayNode `json:"nodes"`
}

type GatewayNode struct {
	ExternalIP string `json:"externalIP"`
	InternalIP string `json:"InternalIP"`
	Name       string `json:"name"`
}

type NFS struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

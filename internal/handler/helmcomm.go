package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	r1 "goodrain.com/cloud-adaptor/internal/usecase"
	"net/http"
)

// HelmHandler -

type HelmHandler struct {
}

// NewHelmHandler -

func NewHelmHandler() *HelmHandler {
	return &HelmHandler{}
}

// GetHelmCommand returns the information of .
//
// swagger:route POST /enterprise-server/api/v1/helm/chart
//
// GetHelmCommand request
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:command
// 400: body:Response
// 500: body:Response

func (h *HelmHandler) GetHelmCommand(c *gin.Context) {
	// 声明接收的变量
	var (
		ci                 r1.ChartInfo
		helmCommand        string
		imageHubCom        string
		etcCom             string
		storageCom         string
		RWXCom             string
		RWOCom             string
		dbCom              string
		uidb               string
		regiondb           string
		nodesForChaosCom   string
		nodesForGatewayCom string
		cloud              string
	)

	// 解析request-body到结构体chartInfo
	if err := c.ShouldBindJSON(&ci); err != nil {
		// 返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 解析数据生成命令
	fmt.Println("chartInfo:", ci)
	// 拼接helm命令  修改value.yaml参数
	// 对外网关
	gatewayIngressIPCom := fmt.Sprintf("--set Cluster.gatewayIngressIPs=%s ", ci.GatewayIngressIPs)
	// 高可用安装
	EnableHACom := fmt.Sprintf("--set Cluster.enableHA=%v ", ci.EnableHA)
	// 镜像仓库
	imh := ci.ImageHub
	if imh != nil && imh.Enable {
		imageHubCom = fmt.Sprintf("--set Cluster.imageHub.enable=%v --set Cluster.imageHub.domain=%v --set Cluster.imageHub.namespace=%v --set Cluster.imageHub.username=%v --set Cluster.imageHub.password=%v ",
			imh.Enable, imh.Domain, imh.Namespace, imh.Username, imh.Password)
	} else {
		fmt.Println("选择内置镜像仓库")
	}
	// Etcd
	etcd := ci.Etcd
	if etcd != nil && etcd.Enable {
		eipCom := ""
		for i, eip := range etcd.Endpoints {
			eipCom += fmt.Sprintf("--set Cluster.etcd.endpoints[%v]=%v ", i, eip.Ip)
		}
		etcCom = fmt.Sprintf("--set Cluster.etcd.enable=%v %s--set Cluster.etcd.secretName=%v ", etcd.Enable, eipCom, etcd.SecretName)
	} else {
		fmt.Println("选择内置etcd")
	}
	// Estorage
	storage := ci.Estorage
	if storage != nil && storage.Enable {
		switch ci.DockingType {
		case "aliyun":
			if storage.RWX.Enable {
				RWXCom = fmt.Sprintf("--set Cluster.RWX.enable=%v --set Cluster.RWX.type=%v --set Cluster.RWX.config.server=%v ", storage.RWX.Enable, ci.DockingType, storage.RWX.Config.Server)
			}
			if storage.RWO.Enable {
				RWOCom = fmt.Sprintf("--set Cluster.RWO.enable=%v --set Cluster.RWO.storageClassName=%v ", storage.RWO.Enable, storage.RWO.StorageClassName)
			}
			storageCom = RWXCom + RWOCom
		case "nfs":
			storageCom = fmt.Sprintf("--set nfs-client-provisioner.childChart.enable=%v --set nfs-client-provisioner.nfs.server=%v --set nfs-client-provisioner.nfs.path=%v --set Cluster.RWX.enable=%v --set Cluster.RWX.type=%v --set Cluster.RWX.config.storageClassName=%v --set Cluster.RWO.enable=%v --set Cluster.RWO.storageClassName=%v ", true, storage.NFS.Server, storage.NFS.Path, true, ci.DockingType, "nfs-client", true, "nfs-client")
		default:
			if storage.RWX.Enable {
				RWXCom = fmt.Sprintf("--set Cluster.RWX.enable=%v --set Cluster.RWX.config.storageClassName=%v ", storage.RWX.Enable, storage.RWX.Config.StorageClassName)
			}
			if storage.RWO.Enable {
				RWOCom = fmt.Sprintf("--set Cluster.RWO.enable=%v --set Cluster.RWO.storageClassName=%v ", storage.RWO.Enable, storage.RWO.StorageClassName)
			}
			storageCom = RWXCom + RWOCom
		}
	} else {
		fmt.Println("使用内置存储")
	}
	// Database
	dbs := ci.Database
	if dbs != nil && dbs.Enable {
		if dbs.RegionDatabase.Enable {
			// 控制台数据库
			uidb = fmt.Sprintf("--set Cluster.uiDatabase.host=%s --set Cluster.uiDatabase.port=%s --set Cluster.uiDatabase.username=%s --set Cluster.uiDatabase.password=%s --set Cluster.uiDatabase.name=%s --set Cluster.uiDatabase.enable=%v ",
				dbs.RegionDatabase.Host, dbs.RegionDatabase.Port, dbs.RegionDatabase.Username, dbs.RegionDatabase.Password, dbs.RegionDatabase.Dbname, dbs.RegionDatabase.Enable,
			)
			// region数据库
			regiondb = fmt.Sprintf("--set Cluster.regionDatabase.host=%s --set Cluster.regionDatabase.port=%s --set Cluster.regionDatabase.username=%s --set Cluster.regionDatabase.password=%s --set Cluster.regionDatabase.name=%s --set Cluster.regionDatabase.enable=%v ",
				dbs.RegionDatabase.Host, dbs.RegionDatabase.Port, dbs.RegionDatabase.Username, dbs.RegionDatabase.Password, dbs.RegionDatabase.Dbname, dbs.RegionDatabase.Enable)
		}
		dbCom = uidb + regiondb
	} else {
		fmt.Println(" 使用内置数据库")
	}
	// nodesForChaos
	if ci.NodesForChaos != nil && ci.NodesForChaos.Enable {
		for i, node := range ci.NodesForChaos.Nodes {
			nodesForChaosCom += fmt.Sprintf("--set Cluster.nodesForChaos[%v].name=%v ", i, node.Name)
		}
	} else {
		fmt.Println("使用默认构建节点")
	}

	// nodesForGateway
	if ci.NodesForGateway != nil && ci.NodesForGateway.Enable {
		for i, node := range ci.NodesForGateway.Nodes {
			nodesForGatewayCom += fmt.Sprintf("--set Cluster.nodesForGateway[%v].name=%v --set Cluster.nodesForGateway[%v].externalIP=%v --set Cluster.nodesForGateway[%v].internalIP=%v ", i, node.Name, i, node.ExternalIP, i, node.InternalIP)
		}
	} else {
		fmt.Println("使用默认构建节点")
	}

	// 是否安装控制台
	appui := fmt.Sprintf("--set Component.rbd_app_ui.enable=%v ", ci.AppUI)

	// 获取token标识
	tokenCom := fmt.Sprintf("--set operator.env[0].name=HELM_TOKEN --set operator.env[0].value=%s ", ci.Token)

	// 获取企业id
	eidCom := fmt.Sprintf("--set operator.env[1].name=ENTERPRISE_ID --set operator.env[1].value=%s ", ci.EID)

	// 获取控制台域名
	domainCom := fmt.Sprintf("--set operator.env[2].name=CONSOLE_DOMAIN --set operator.env[2].value=%s/console/enterprise/helm/region_info ", ci.Domain)

	// 获取云服务
	cloud = ""
	if ci.CloudServer != "" {
		cloud = fmt.Sprintf("--set operator.env[3].name=CLOUD_SERVER --set operator.env[3].value=%s", ci.CloudServer)
	}
	repoCom := "kubectl create namespace rbd-system & helm repo add rainbond https://openchart.goodrain.com/goodrain/rainbond & helm repo update & helm install "
	commTail := "rainbond rainbond/rainbond-cluster -n rbd-system "
	// 拼接所有命令
	helmCommand = fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", repoCom, commTail, gatewayIngressIPCom, EnableHACom, imageHubCom, etcCom, storageCom, dbCom, nodesForChaosCom, nodesForGatewayCom, tokenCom, eidCom, domainCom, appui, cloud)
	fmt.Println("helmCommand:", helmCommand)

	c.JSON(http.StatusOK, gin.H{"status": 200, "command": helmCommand})
}

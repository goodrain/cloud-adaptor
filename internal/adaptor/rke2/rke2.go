package rke2

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"goodrain.com/cloud-adaptor/internal/datastore"
	"goodrain.com/cloud-adaptor/internal/model"
	"strings"
	"time"
)

func InitConn(rke2Server *model.RKE2Nodes) (conn *ssh.Client, err error) {
	// 配置SSH客户端参数
	config := &ssh.ClientConfig{
		User: rke2Server.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(rke2Server.Pass),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// 尝试连接目标主机
	conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", rke2Server.Host, rke2Server.Port), config)
	if err != nil {
		logrus.Errorf("Failed to dial: %s", err)
		return nil, err
	}
	return conn, err
}
func UninstallRKE2Node(rke2Server *model.RKE2Nodes) {
	conn, err := InitConn(rke2Server)

	if err != nil {
		logrus.Errorf("Failed to dial: %s", err)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err)
		return
	}

	defer session.Close()
	err = session.Run("/usr/bin/rke2-uninstall.sh")
	if err != nil {
		logrus.Errorf("Failed to execute rke2-uninstall.sh command: %s", err)
		return
	}
}

func InstallRKE2Cluster(cluster *model.RKECluster, rke2Server *model.RKE2Nodes) error {
	conn, err := InitConn(rke2Server)

	if err != nil {
		logrus.Errorf("Failed to dial: %s", err)
		return err
	}
	defer func(conn *ssh.Client) {
		err := conn.Close()
		if err != nil {
			logrus.Errorf("错误")
		}
	}(conn)

	// 如果是agent 或者其他server，那么要写入一个 host 配置
	if cluster == nil {
		var node model.RKE2Nodes
		err2 := datastore.GetGDB().First(&node, "cluster_id = ? AND stats = 'running' AND role = 'server'", rke2Server.ClusterID).Error
		if err2 != nil {
			return err2
		}

		session, err3 := conn.NewSession()
		if err3 != nil {
			return err3
		}
		err3 = session.Run(fmt.Sprintf("echo \"%s    goodrain.rke2\" >> /etc/hosts", node.Host))
		if err3 != nil {
			return err3
		}
	}

	rke2Server.Stats = "installing"
	datastore.GetGDB().Save(rke2Server)
	// 安装rke2
	return installRKE2(conn, rke2Server, cluster)
}

func installRKE2(conn *ssh.Client, rke2Server *model.RKE2Nodes, cluster *model.RKECluster) error {
	session, err := conn.NewSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err)
		return err
	}

	err = session.Run("curl -sfL https://get.rainbond.com/rke2-install.sh | INSTALL_RKE2_VERSION=v1.25.16+rke2r1 INSTALL_RKE2_MIRROR=cn INSTALL_RKE2_TYPE=\"" + rke2Server.Role + "\" sh -")
	if err != nil {
		logrus.Errorf("Failed to execute installRKE2 command: %s", err)
		return err
	}
	// 步骤二：创建配置文件
	err = saveConfig(conn, rke2Server, cluster)
	if err != nil {
		return err
	}

	// 步骤三：启动
	err = start(conn, rke2Server.Role)
	if err != nil {
		return err
	}
	//主节点才会去获取kube config
	if cluster != nil {
		// 步骤四：保存kubeconfig文件
		kubeconfig, err := saveKubeconfig(conn)
		if err != nil {
			return err
		}
		cluster.KubeConfig = strings.Replace(kubeconfig, "127.0.0.1", rke2Server.Host, -1)
		cluster.APIURL = "https://" + rke2Server.Host + ":6443"
		datastore.GetGDB().Save(cluster)

		// 步骤五： 自动创建rbd-system命名空间
		session2, err := conn.NewSession()
		if err != nil {
			logrus.Errorf("Failed to create session: %s", err)
			return err
		}
		defer session2.Close()
		err = session.Run("kubectl create ns rbd-system")
		if err != nil {
			logrus.Errorf("Failed to exec create ns: %s", err)
			return err
		}
	}
	return nil
}

func saveKubeconfig(conn *ssh.Client) (string, error) {
	session, err := conn.NewSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err)
		return "", err
	}

	output, err := session.CombinedOutput("cat /etc/rancher/rke2/rke2.yaml")
	if err != nil {
		logrus.Errorf("Failed to execute saveKubeconfig command: %s", err)
		return "", err
	}
	return string(output), nil
}

// saveConfig 保存并且新建配置文件
func saveConfig(conn *ssh.Client, rke2Server *model.RKE2Nodes, cluster *model.RKECluster) error {
	session, err := conn.NewSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err)
		return err
	}

	defer session.Close()
	var staticConfig = `token: goodrain:rke2
node-external-ip: ` + rke2Server.Host + `
system-default-registry: "registry.cn-hangzhou.aliyuncs.com"
disable: 
  - rke2-ingress-nginx
  - rke2-metrics-server
tls-san:
  - goodrain.rke2`

	if cluster == nil {
		staticConfig += "\nserver: https://goodrain.rke2:9345"
	}
	err = session.Run(fmt.Sprintf("mkdir -p /etc/rancher/rke2/config.yaml.d/; echo \"%s\" > /etc/rancher/rke2/config.yaml; cd /etc/rancher/rke2/config.yaml.d; echo \"%s\" > static.yaml", rke2Server.ConfigFile, staticConfig))
	if err != nil {
		logrus.Errorf("Failed to execute saveConfig command: %s", err)
		return err
	}
	return nil
}

func start(conn *ssh.Client, InstallRke2Type string) error {
	cmd := "systemctl enable rke2-server.service; systemctl start rke2-server.service; cp /var/lib/rancher/rke2/bin/kubectl /usr/local/bin/kubectl; mkdir .kube; cp /etc/rancher/rke2/rke2.yaml .kube/config; "

	if InstallRke2Type == "agent" {
		cmd = "systemctl enable rke2-agent.service; systemctl start rke2-agent.service"
	}
	session, err := conn.NewSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err)
		return err
	}

	defer session.Close()

	err = session.Run(cmd)
	if err != nil {
		logrus.Errorf("Failed to execute rke2 start command: %s", err)
		return err
	}
	return nil
}

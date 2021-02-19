// RAINBOND, Application Management Platform
// Copyright (C) 2020-2020 Goodrain Co., Ltd.

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

package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/api/models"
	"goodrain.com/cloud-adaptor/util/ginutil"
	"k8s.io/client-go/util/homedir"
)

//SystemHandler -
type SystemHandler struct {
	DB *gorm.DB `inject:""`
}

// NewSystemHandler new system handler
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

//Backup backup all data
func (s SystemHandler) Backup(ctx *gin.Context) {
	//backup dir
	backupTmpPath := "/tmp/backup/"
	defer func() {
		os.RemoveAll(backupTmpPath)
	}()
	os.MkdirAll(backupTmpPath, 0755)
	//backup db data
	var result models.BackupListModelData
	s.DB.Table(s.DB.NewScope(&models.CloudAccessKey{}).TableName()).Scan(&result.CloudAccessKeys)
	s.DB.Table(s.DB.NewScope(&models.CreateKubernetesTask{}).TableName()).Scan(&result.CreateKubernetesTasks)
	s.DB.Table(s.DB.NewScope(&models.InitRainbondTask{}).TableName()).Scan(&result.InitRainbondTasks)
	s.DB.Table(s.DB.NewScope(&models.TaskEvent{}).TableName()).Scan(&result.TaskEvents)
	s.DB.Table(s.DB.NewScope(&models.UpdateKubernetesTask{}).TableName()).Scan(&result.UpdateKubernetesTasks)
	s.DB.Table(s.DB.NewScope(&models.CustomCluster{}).TableName()).Scan(&result.CustomClusters)
	s.DB.Table(s.DB.NewScope(&models.RKECluster{}).TableName()).Scan(&result.RKEClusters)
	s.DB.Table(s.DB.NewScope(&models.RainbondClusterConfig{}).TableName()).Scan(&result.RainbondClusterConfigs)
	data, err := json.Marshal(result)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	if err := ioutil.WriteFile(path.Join(backupTmpPath, "cloudadaptor-db.json"), data, 0755); err != nil {
		logrus.Errorf("write backup db file failure %s", err.Error())
		ginutil.JSON(ctx, nil, err)
		return
	}
	//backup rke data
	configDir := "/tmp"
	if os.Getenv("CONFIG_DIR") != "" {
		configDir = os.Getenv("CONFIG_DIR")
	}
	rkeDir := path.Join(configDir, "rke")
	_, err = os.Stat(rkeDir)
	if err != nil {
		os.MkdirAll(rkeDir, 0755)
	}
	tarPackgeFile := path.Join(backupTmpPath, "rke.tar.gz")
	cmd := exec.Command("tar", "-czf", tarPackgeFile, "./")
	cmd.Dir = rkeDir
	if err := cmd.Run(); err != nil {
		logrus.Errorf("write backup rke file failure %s", err.Error())
		ginutil.JSON(ctx, nil, err)
		return
	}
	//backup ssh key
	tarSSHPackgeFile := path.Join(backupTmpPath, "ssh.tar.gz")
	scmd := exec.Command("tar", "-czf", tarSSHPackgeFile, "./id_rsa", "./id_rsa.pub")
	scmd.Dir = path.Join(homedir.HomeDir(), ".ssh")
	if err := scmd.Run(); err != nil {
		logrus.Errorf("write backup ssh file failure %s", err.Error())
		ginutil.JSON(ctx, nil, err)
		return
	}
	// create packege
	backupTmpFile := "/tmp/cloud_adaptor_data.tar.gz"
	packeCmd := exec.Command("tar", "-czf", backupTmpFile, "./")
	packeCmd.Dir = backupTmpPath
	if err := packeCmd.Run(); err != nil {
		logrus.Errorf("write backup file failure %s", err.Error())
		ginutil.JSON(ctx, nil, err)
		return
	}
	filedata, err := ioutil.ReadFile(backupTmpFile)
	if err != nil {
		logrus.Errorf("read backup file failure %s", err.Error())
		ginutil.JSON(ctx, nil, err)
		return
	}
	// Backup files are usually small and read directly into memory.
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename=cloud_adaptor_data.tar.gz")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(filedata)))
	ctx.Writer.Write(filedata)
}

//Recover all data
func (s SystemHandler) Recover(c *gin.Context) {
	recoverPath := "/tmp/recover"
	_, err := os.Stat(recoverPath)
	if err == nil || !os.IsNotExist(err) {
		ginutil.JSON(c, "recover is running", nil)
		return
	}
	f, err := c.FormFile("file")
	if err != nil {
		ginutil.JSON(c, nil, err)
		return
	}
	defer func() {
		os.RemoveAll(recoverPath)
	}()
	os.Mkdir(recoverPath, 0755)

	dst := path.Join(recoverPath, f.Filename)
	err = c.SaveUploadedFile(f, dst)
	if err != nil {
		logrus.Errorf("save upload backup file failure %s", err.Error())
		ginutil.JSON(c, nil, err)
		return
	}
	if err := exec.Command("tar", "-xzf", dst, "-C", recoverPath).Run(); err != nil {
		logrus.Errorf("untar backup file failure %s", err.Error())
		ginutil.JSON(c, nil, err)
		return
	}
	// recover rke data
	rkePath := path.Join(recoverPath, "rke.tar.gz")
	rkeFile, err := os.Stat(rkePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("check rke backup file failure %s", err.Error())
		ginutil.JSON(c, nil, err)
		return
	}
	if rkeFile != nil {
		logrus.Infof("start recover rke backup data")
		configDir := "/tmp"
		if os.Getenv("CONFIG_DIR") != "" {
			configDir = os.Getenv("CONFIG_DIR")
		}
		rkeDir := path.Join(configDir, "rke")
		os.MkdirAll(rkeDir, 0755)
		if err := exec.Command("tar", "-xzf", rkePath, "-C", rkeDir).Run(); err != nil {
			logrus.Errorf("recover rke data failure %s", err.Error())
		}
		logrus.Infof("recover rke backup data success")
	}
	// recover ssh data
	sshPath := path.Join(recoverPath, "ssh.tar.gz")
	sshFile, err := os.Stat(sshPath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("check rke backup file failure %s", err.Error())
		ginutil.JSON(c, nil, err)
		return
	}
	if sshFile != nil {
		logrus.Infof("start recover ssh backup data")
		sshDir := path.Join(homedir.HomeDir(), ".ssh")
		os.MkdirAll(sshDir, 0700)
		if err := exec.Command("tar", "-xzf", sshPath, "-C", sshDir).Run(); err != nil {
			logrus.Errorf("recover ssh data failure %s", err.Error())
		}
		logrus.Infof("recover ssh backup data success")
	}
	// recover db data
	bytes, err := ioutil.ReadFile(path.Join(recoverPath, "cloudadaptor-db.json"))
	if err != nil {
		logrus.Errorf("read db backup file failure ", err.Error())
	} else {
		logrus.Infof("start recover db backup data")
		tx := s.DB.Begin()
		if func() error {
			var data models.BackupListModelData
			err = json.Unmarshal(bytes, &data)
			if err != nil {
				logrus.Errorf("unmarshal db backup file failure ", err.Error())
			}
			if err := tx.Delete(&models.CloudAccessKey{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.CreateKubernetesTask{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.InitRainbondTask{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.UpdateKubernetesTask{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.TaskEvent{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.CustomCluster{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.RKECluster{}).Error; err != nil {
				return err
			}
			for _, accessKey := range data.CloudAccessKeys {
				if err := tx.Create(&accessKey).Error; err != nil {
					return fmt.Errorf("recover accessKey failure %s", err.Error())
				}
			}
			for _, createTask := range data.CreateKubernetesTasks {
				if err := tx.Create(&createTask).Error; err != nil {
					return fmt.Errorf("recover createTask failure %s", err.Error())
				}
			}
			for _, initTask := range data.InitRainbondTasks {
				if err := tx.Create(&initTask).Error; err != nil {
					return fmt.Errorf("recover initTask failure %s", err.Error())
				}
			}
			for _, taskEvent := range data.TaskEvents {
				if err := tx.Create(&taskEvent).Error; err != nil {
					return fmt.Errorf("recover taskEvent failure %s", err.Error())
				}
			}

			for _, updateTask := range data.UpdateKubernetesTasks {
				if err := tx.Create(&updateTask).Error; err != nil {
					return fmt.Errorf("recover updateTask failure %s", err.Error())
				}
			}
			for _, customCluster := range data.CustomClusters {
				if err := tx.Create(&customCluster).Error; err != nil {
					return fmt.Errorf("recover customCluster failure %s", err.Error())
				}
			}
			for _, rkeCluster := range data.RKEClusters {
				if err := tx.Create(&rkeCluster).Error; err != nil {
					return fmt.Errorf("recover rkeCluster failure %s", err.Error())
				}
			}
			for _, rcc := range data.RainbondClusterConfigs {
				if err := tx.Create(&rcc).Error; err != nil {
					return fmt.Errorf("recover rainbondClusterConfigs failure %s", err.Error())
				}
			}
			logrus.Infof("recover db backup data success")
			return nil
		}(); err != nil {
			tx.Rollback()
			logrus.Errorf("recover db data failure %s", err.Error())
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			logrus.Errorf("recover db data failure %s", err.Error())
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

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

package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/util/homedir"
)

// GenerateKey -
func GenerateKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return private, &private.PublicKey, nil

}

// EncodePrivateKey -
func EncodePrivateKey(private *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(private),
		Type:  "RSA PRIVATE KEY",
	})
}

// EncodePublicKey -
func EncodePublicKey(public *rsa.PublicKey) ([]byte, error) {
	publicBytes, err := x509.MarshalPKIXPublicKey(public)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Bytes: publicBytes,
		Type:  "PUBLIC KEY",
	}), nil
}

// EncodeSSHKey -
func EncodeSSHKey(public *rsa.PublicKey) ([]byte, error) {
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}

// MakeSSHKeyPair -
func MakeSSHKeyPair() (string, string, error) {

	pkey, pubkey, err := GenerateKey(2048)
	if err != nil {
		return "", "", err
	}

	pub, err := EncodeSSHKey(pubkey)
	if err != nil {
		return "", "", err
	}

	return string(EncodePrivateKey(pkey)), string(pub), nil
}

// GetOrMakeSSHRSA get or make ssh rsa
func GetOrMakeSSHRSA() (string, error) {
	home := homedir.HomeDir()
	if _, err := os.Stat(path.Join(home, ".ssh")); err != nil && os.IsNotExist(err) {
		os.MkdirAll(path.Join(home, ".ssh"), 0700)
	}
	idRsaPath := path.Join(home, ".ssh", "id_rsa")
	idRsaPubPath := path.Join(home, ".ssh", "id_rsa.pub")
	stat, err := os.Stat(idRsaPubPath)
	if os.IsNotExist(err) || stat.IsDir() {
		os.Remove(idRsaPath)
		os.Remove(idRsaPubPath)
		private, pub, err := MakeSSHKeyPair()
		if err != nil {
			return "", fmt.Errorf("create ssh rsa failure %s", err.Error())
		}
		logrus.Infof("init ssh rsa file %s %s ", idRsaPath, idRsaPubPath)
		if err := ioutil.WriteFile(idRsaPath, []byte(private), 0600); err != nil {
			return "", fmt.Errorf("write ssh rsa file failure %s", err.Error())
		}
		if err := ioutil.WriteFile(idRsaPubPath, []byte(pub), 0644); err != nil {
			return "", fmt.Errorf("write ssh rsa pub file failure %s", err.Error())
		}
		return pub, nil
	}
	if err != nil {
		return "", err
	}
	pub, err := ioutil.ReadFile(idRsaPubPath)
	if err != nil {
		return "", fmt.Errorf("read rsa pub file failure %s", err.Error())
	}
	return string(pub), nil
}

// check ssh connection
func CheckSSHConnect(host string, port uint) (bool, error) {
	// 读取私钥文件
	key, err := ioutil.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		return false, bcode.ErrSSHFileNotFond
	}

	// 使用私钥创建一个Signer
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return false, bcode.ErrParseSSH
	}

	// 配置SSH客户端参数
	config := &ssh.ClientConfig{
		User: "docker",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 尝试连接目标主机
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)

	if err != nil {
		return false, bcode.ErrConnect
	}
	defer conn.Close()
	return true, nil
}

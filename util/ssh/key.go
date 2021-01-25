package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/util/homedir"
)

//GenerateKey -
func GenerateKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return private, &private.PublicKey, nil

}

//EncodePrivateKey -
func EncodePrivateKey(private *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(private),
		Type:  "RSA PRIVATE KEY",
	})
}

//EncodePublicKey -
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

//EncodeSSHKey -
func EncodeSSHKey(public *rsa.PublicKey) ([]byte, error) {
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}

//MakeSSHKeyPair -
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

//GetOrMakeSSHRSA get or make ssh rsa
func GetOrMakeSSHRSA() (string, error) {
	home := homedir.HomeDir()
	idRsaPath := path.Join(home, ".ssh", "id_rsa")
	idRsaPubPath := path.Join(home, ".ssh", "id_rsa.pub")
	_, err := os.Stat(idRsaPubPath)
	if os.IsNotExist(err) {
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

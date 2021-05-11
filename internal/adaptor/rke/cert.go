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

package rke

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/pki/cert"
	"github.com/rancher/rke/services"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
)

func rebuildClusterWithRotatedCertificates(ctx context.Context,
	dialersOptions hosts.DialersOptions,
	flags cluster.ExternalFlags, svcOptionData map[string]*v3.KubernetesServicesOptions) (string, string, string, string, map[string]pki.CertificatePKI, error) {
	var APIURL, caCrt, clientCert, clientKey string
	log.Infof(ctx, "Rebuilding Kubernetes cluster with rotated certificates")
	clusterState, err := cluster.ReadStateFile(ctx, cluster.GetStateFilePath(flags.ClusterFilePath, flags.ConfigDir))
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	kubeCluster, err := cluster.InitClusterObject(ctx, clusterState.DesiredState.RancherKubernetesEngineConfig.DeepCopy(), flags, clusterState.DesiredState.EncryptionConfig)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.SetupDialers(ctx, dialersOptions); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err := kubeCluster.TunnelHosts(ctx, flags); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err := cluster.SetUpAuthentication(ctx, kubeCluster, nil, clusterState); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if len(kubeCluster.ControlPlaneHosts) > 0 {
		APIURL = fmt.Sprintf("https://%s:6443", kubeCluster.ControlPlaneHosts[0].Address)
	}
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Key))
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))

	if err := kubeCluster.SetUpHosts(ctx, flags); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	// Save new State
	if err := saveClusterState(ctx, kubeCluster, clusterState); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	// Restarting Kubernetes components
	servicesMap := make(map[string]bool)
	for _, component := range kubeCluster.RotateCertificates.Services {
		servicesMap[component] = true
	}

	if len(kubeCluster.RotateCertificates.Services) == 0 || kubeCluster.RotateCertificates.CACertificates || servicesMap[services.EtcdContainerName] {
		if err := services.RestartEtcdPlane(ctx, kubeCluster.EtcdHosts); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}
	isLegacyKubeAPI, err := cluster.IsLegacyKubeAPI(ctx, kubeCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if isLegacyKubeAPI {
		log.Infof(ctx, "[controlplane] Redeploying controlplane to update kubeapi parameters")
		if _, err := kubeCluster.DeployControlPlane(ctx, svcOptionData, true); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}
	if err := services.RestartControlPlane(ctx, kubeCluster.ControlPlaneHosts); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	allHosts := hosts.GetUniqueHostList(kubeCluster.EtcdHosts, kubeCluster.ControlPlaneHosts, kubeCluster.WorkerHosts)
	if err := services.RestartWorkerPlane(ctx, allHosts); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if kubeCluster.RotateCertificates.CACertificates {
		if err := cluster.RestartClusterPods(ctx, kubeCluster); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}
	return APIURL, caCrt, clientCert, clientKey, kubeCluster.Certificates, nil
}

func saveClusterState(ctx context.Context, kubeCluster *cluster.Cluster, clusterState *cluster.FullState) error {
	var err error
	if err = kubeCluster.UpdateClusterCurrentState(ctx, clusterState); err != nil {
		return err
	}
	// Attempt to store cluster full state to Kubernetes
	for i := 1; i <= 3; i++ {
		err = cluster.SaveFullStateToKubernetes(ctx, kubeCluster, clusterState)
		if err != nil {
			time.Sleep(time.Second * time.Duration(2))
			continue
		}
		break
	}
	if err != nil {
		logrus.Warnf("Failed to save full cluster state to Kubernetes")
	}
	return nil
}

func rotateRKECertificates(ctx context.Context, kubeCluster *cluster.Cluster, flags cluster.ExternalFlags, rkeFullState *cluster.FullState) (*cluster.FullState, error) {
	log.Infof(ctx, "Rotating Kubernetes cluster certificates")
	currentCluster, err := kubeCluster.GetClusterState(ctx, rkeFullState)
	if err != nil {
		return nil, err
	}
	if currentCluster == nil {
		return nil, fmt.Errorf("Failed to rotate certificates: can't find old certificates")
	}
	currentCluster.RotateCertificates = kubeCluster.RotateCertificates
	if !kubeCluster.RotateCertificates.CACertificates {
		caCertPKI, ok := rkeFullState.CurrentState.CertificatesBundle[pki.CACertName]
		if !ok {
			return nil, fmt.Errorf("Failed to rotate certificates: can't find CA certificate")
		}
		caCert := caCertPKI.Certificate
		if caCert == nil {
			return nil, fmt.Errorf("Failed to rotate certificates: CA certificate is nil")
		}
		certPool := x509.NewCertPool()
		certPool.AddCert(caCert)
		if _, err := caCert.Verify(x509.VerifyOptions{Roots: certPool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); err != nil {
			return nil, fmt.Errorf("Failed to rotate certificates: CA certificate is invalid, please use the --rotate-ca flag to rotate CA certificate, error: %v", err)
		}
	}
	if err := cluster.RotateRKECertificates(ctx, currentCluster, flags, rkeFullState); err != nil {
		return nil, err
	}
	rkeFullState.DesiredState.RancherKubernetesEngineConfig = &kubeCluster.RancherKubernetesEngineConfig
	return rkeFullState, nil
}

// GenerateRKECSRs -
func GenerateRKECSRs(ctx context.Context, rkeConfig *v3.RancherKubernetesEngineConfig, flags cluster.ExternalFlags) error {
	log.Infof(ctx, "Generating Kubernetes cluster CSR certificates")
	if len(flags.CertificateDir) == 0 {
		flags.CertificateDir = cluster.GetCertificateDirPath(flags.ClusterFilePath, flags.ConfigDir)
	}

	certBundle, err := pki.ReadCSRsAndKeysFromDir(flags.CertificateDir)
	if err != nil {
		return err
	}

	// initialze the cluster object from the config file
	kubeCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, "")
	if err != nil {
		return err
	}

	// Generating csrs for kubernetes components
	if err := pki.GenerateRKEServicesCSRs(ctx, certBundle, kubeCluster.RancherKubernetesEngineConfig); err != nil {
		return err
	}
	return pki.WriteCertificates(kubeCluster.CertificateDir, certBundle)
}

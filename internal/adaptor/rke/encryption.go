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
	"fmt"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/pki/cert"
	v3 "github.com/rancher/rke/types"
)

//RotateEncryptionKey -
func RotateEncryptionKey(
	ctx context.Context,
	rkeConfig *v3.RancherKubernetesEngineConfig,
	dialersOptions hosts.DialersOptions,
	flags cluster.ExternalFlags,
) (string, string, string, string, map[string]pki.CertificatePKI, error) {
	log.Infof(ctx, "Rotating cluster secrets encryption key")

	var APIURL, caCrt, clientCert, clientKey string

	stateFilePath := cluster.GetStateFilePath(flags.ClusterFilePath, flags.ConfigDir)
	rkeFullState, _ := cluster.ReadStateFile(ctx, stateFilePath)

	// We generate the first encryption config in ClusterInit, to store it ASAP. It's written to the DesiredState
	stateEncryptionConfig := rkeFullState.DesiredState.EncryptionConfig
	// if CurrentState has EncryptionConfig, it means this is NOT the first time we enable encryption, we should use the _latest_ applied value from the current cluster
	if rkeFullState.CurrentState.EncryptionConfig != "" {
		stateEncryptionConfig = rkeFullState.CurrentState.EncryptionConfig
	}

	kubeCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, stateEncryptionConfig)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if kubeCluster.IsEncryptionCustomConfig() {
		return APIURL, caCrt, clientCert, clientKey, nil, fmt.Errorf("can't rotate encryption keys: Key Rotation is not supported with custom configuration")
	}
	if !kubeCluster.IsEncryptionEnabled() {
		return APIURL, caCrt, clientCert, clientKey, nil, fmt.Errorf("can't rotate encryption keys: Encryption Configuration is disabled")
	}

	kubeCluster.Certificates = rkeFullState.DesiredState.CertificatesBundle
	if err := kubeCluster.SetupDialers(ctx, dialersOptions); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.TunnelHosts(ctx, flags); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if len(kubeCluster.ControlPlaneHosts) > 0 {
		APIURL = fmt.Sprintf("https://%s:6443", kubeCluster.ControlPlaneHosts[0].Address)
	}
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Key))
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))

	err = kubeCluster.RotateEncryptionKey(ctx, rkeFullState)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	// make sure we have the latest state
	rkeFullState, _ = cluster.ReadStateFile(ctx, stateFilePath)

	log.Infof(ctx, "Reconciling cluster state")
	if err := kubeCluster.ReconcileDesiredStateEncryptionConfig(ctx, rkeFullState); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	log.Infof(ctx, "Cluster secrets encryption key rotated successfully")
	return APIURL, caCrt, clientCert, clientKey, kubeCluster.Certificates, nil
}

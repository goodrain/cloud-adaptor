// RAINBOND, Application Management Platform
// Copyright (C) 2021-2021 Goodrain Co., Ltd.

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

package versionutil

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/version"
)

// CheckVersion Check whether the k8s version is between 1.16 and 1.19
func CheckVersion(kubernetesVersion string) bool {
	clusterVersion, err := version.ParseGeneric(kubernetesVersion)
	if err != nil {
		logrus.Errorf("parse kubernetes version %s failed", kubernetesVersion)
		return false
	}
	minK8sVersion, _ := version.ParseGeneric("v1.19.0")
	maxK8sVersion, _ := version.ParseGeneric("v1.26.0")
	return clusterVersion.AtLeast(minK8sVersion) && clusterVersion.LessThan(maxK8sVersion)
}

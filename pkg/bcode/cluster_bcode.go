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

package bcode

var (
	//ErrorProviderNotSupport provider not support
	ErrorProviderNotSupport = newByMessage(400, 7001, "provider not support")
	//ErrorNotFoundAccessKey not found access key
	ErrorNotFoundAccessKey = newByMessage(400, 7002, "not found access key with this enterprise")
	//ErrorNotSetAccessKey not set access key
	ErrorNotSetAccessKey = newByMessage(200, 7003, "not found access key with this enterprise")
	//ErrorAccessKeyNotMatch access key and secret not match
	ErrorAccessKeyNotMatch = newByMessage(403, 7004, "access key and secret not match")
	//ErrorLastTaskNotComplete last task not complete
	ErrorLastTaskNotComplete = newByMessage(400, 7005, "last task can not complete")
	//ErrorKubeAPI kube api connection error
	ErrorKubeAPI = newByMessage(400, 7006, "kube api connection error")
	//ErrorAliyunError error aliyun api error
	ErrorAliyunError = newByMessage(400, 7007, "aliyun api failure")
	//ErrorClusterRoleNotExist aliyun list cluster role
	ErrorClusterRoleNotExist = newByMessage(400, 7008, "The role not exists: role/aliyuncsdefaultrole")
	//ErrClusterNodeEmpty cluster node empty error
	ErrClusterNodeEmpty = newByMessage(400, 7009, "RKE node must define")
	//ErrClusterNodeRoleMiss cluster node role miss
	ErrClusterNodeRoleMiss = newByMessage(400, 7010, "RKE node role miss")
	//ErrETCDNodeNotOddNumer -
	ErrETCDNodeNotOddNumer = newByMessage(400, 7011, "RKE etcd node must odd number")
	//ErrClusterNodeIPInvalid -
	ErrClusterNodeIPInvalid = newByMessage(400, 7012, "RKE node ip is invalid")
	//ErrClusterNodePortInvalid -
	ErrClusterNodePortInvalid = newByMessage(400, 7013, "RKE node ssh port is invalid")
	//ErrKubeConfigCannotEmpty -
	ErrKubeConfigCannotEmpty = newByMessage(400, 7014, "kube config can not be empty")

	//ErrClusterNotAllowDelete -
	ErrClusterNotAllowDelete = newByMessage(400, 7015, "cluster can not be delete")
	//ErrNotSupportReInstall -
	ErrNotSupportReInstall = newByMessage(400, 7016, "cluster can not support reinstall")

	//ErrNotSupportUpdateKubernetes -
	ErrNotSupportUpdateKubernetes = newByMessage(400, 7017, "cluster can not support update kubernetes")

	//ErrConfigInvalid -
	ErrConfigInvalid = newByMessage(400, 7018, "rainbond cluster config is invalid")

	//ErrorGetRegionStatus -
	ErrorGetRegionStatus = newByMessage(400, 7019, "can not get region status")

	ErrIncorrectRKEConfig = newByMessage(400, 7020, "the rke configuration format is incorrect")
	ErrRKEConfigLost      = newByMessage(404, 7021, "rancher kubernetes engine configuration lost")
)

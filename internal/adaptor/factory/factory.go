// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

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

package factory

import (
	"fmt"

	"goodrain.com/cloud-adaptor/internal/adaptor"
	"goodrain.com/cloud-adaptor/internal/adaptor/ack"
	"goodrain.com/cloud-adaptor/internal/adaptor/custom"
	"goodrain.com/cloud-adaptor/internal/adaptor/rke"
)

//ErrorNotSupport not support adaptor
var ErrorNotSupport = fmt.Errorf("adaptor type not support")
var defaultCloudFactory = &cloudFactory{}

//CloudFactory cloud adaptor factory
type CloudFactory interface {
	GetAdaptor(adaptor, accessKeyID, accessKeySecret string) (adaptor.CloudAdaptor, error)
	GetRainbondClusterAdaptor(adaptor, accessKeyID, accessKeySecret string) (adaptor.RainbondClusterAdaptor, error)
}

//cloudFactory -
type cloudFactory struct {
}

//GetCloudFactory get cloud factory
func GetCloudFactory() CloudFactory {
	return defaultCloudFactory
}

func (f *cloudFactory) GetAdaptor(adaptorType, accessKeyID, accessKeySecret string) (adaptor.CloudAdaptor, error) {
	switch adaptorType {
	case "ack":
		return ack.Create(accessKeyID, accessKeySecret)
	default:
		return nil, ErrorNotSupport
	}
}

func (f *cloudFactory) GetRainbondClusterAdaptor(adaptorType, accessKeyID, accessKeySecret string) (adaptor.RainbondClusterAdaptor, error) {
	switch adaptorType {
	case "ack":
		return ack.Create(accessKeyID, accessKeySecret)
	case "rke":
		return rke.Create()
	case "custom":
		return custom.Create()
	default:
		return nil, ErrorNotSupport
	}
}

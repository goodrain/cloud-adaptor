module goodrain.com/cloud-adaptor

go 1.13

require (
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.94
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/inject v0.0.0-20180706035515-f23751cae28b
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/structtag v0.0.0-20150214074306-217e25fb9691 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/goodrain/rainbond-operator v1.3.1-0.20210316113733-75a870bf2a51
	github.com/google/subcommands v1.2.0 // indirect
	github.com/google/wire v0.5.0
	github.com/jinzhu/copier v0.2.8 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/nsqio/go-nsq v1.0.8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	github.com/rancher/norman v0.0.0-20200520181341-ab75acb55410 // indirect
	github.com/rancher/rke v1.2.0-rc9.0.20210106190313-91aed199f04c
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57 // indirect
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.7.0
)

replace (
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190808172531-150530564a14
	github.com/googleapis/gnostic/OpenAPIv2 => github.com/googleapis/gnostic/openapiv2 v0.4.1
	k8s.io/api => k8s.io/api v0.20.1
	k8s.io/client-go => k8s.io/client-go v0.20.1
	k8s.io/kubectl => k8s.io/kubectl v0.20.1
)

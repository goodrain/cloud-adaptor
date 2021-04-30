module goodrain.com/cloud-adaptor

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.94
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/validator/v10 v10.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/goodrain/rainbond-operator v1.3.1-0.20210316113733-75a870bf2a51
	github.com/google/wire v0.5.0
	github.com/helm/helm v2.17.0+incompatible
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/nsqio/go-nsq v1.0.8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	github.com/rancher/norman v0.0.0-20200520181341-ab75acb55410 // indirect
	github.com/rancher/rke v1.2.0-rc9.0.20210106190313-91aed199f04c
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.6.7
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/ugorji/go v1.2.5 // indirect
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210423184538-5f58ad60dda6 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.0.5
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.7
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/helm v2.17.0+incompatible
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190808172531-150530564a14
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8
	github.com/googleapis/gnostic/OpenAPIv2 => github.com/googleapis/gnostic/openapiv2 v0.4.1
	k8s.io/api => k8s.io/api v0.20.1
	k8s.io/client-go => k8s.io/client-go v0.20.1
	k8s.io/kubectl => k8s.io/kubectl v0.20.1
)

module goodrain.com/cloud-adaptor

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.94
	github.com/devfeel/mapper v0.7.5
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/validator/v10 v10.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/goodrain/rainbond v0.0.0-00010101000000-000000000000
	github.com/goodrain/rainbond-operator v1.3.1-0.20210401055914-f8fe4bf89a21
	github.com/google/wire v0.5.0
	github.com/helm/helm v2.17.0+incompatible
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/nsqio/go-nsq v1.0.8
	github.com/pkg/errors v0.9.1
	github.com/rancher/rancher/pkg/apis v0.0.0-20210507220919-8c014efa8531
	github.com/rancher/rke v1.3.15
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.6.7
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/ugorji/go v1.2.5 // indirect
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.0.5
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.7
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/helm v2.17.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0-beta.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v20.10.0-rc2+incompatible
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8
	github.com/goodrain/rainbond => github.com/glyasai/rainbond v0.0.0-20210508110948-5d37bb0288f7
	// github.com/goodrain/rainbond => /Users/abewang/Goodrain/workspaces/rainbond
	github.com/googleapis/gnostic/OpenAPIv2 => github.com/googleapis/gnostic/openapiv2 v0.4.1
	k8s.io/api => k8s.io/api v0.20.1
	k8s.io/client-go => k8s.io/client-go v0.20.1
	k8s.io/kubectl => k8s.io/kubectl v0.20.1
)

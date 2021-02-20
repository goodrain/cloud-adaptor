package version

import (
	"os"
	"strings"
)

//RainbondRegionVersion rainbond region install version
var RainbondRegionVersion = "v5.3.0-release"

//OperatorVersion operator image tag
var OperatorVersion = "v2.0.0"

//InstallImageRepo install image repo
var InstallImageRepo = "registry.cn-hangzhou.aliyuncs.com/goodrain"

func init() {
	if os.Getenv("INSTALL_IMAGE_REPO") != "" {
		InstallImageRepo = os.Getenv("INSTALL_IMAGE_REPO")
	}
	if os.Getenv("RAINBOND_VERSION") != "" {
		RainbondRegionVersion = os.Getenv("RAINBOND_VERSION")
	}
	if strings.HasSuffix(InstallImageRepo, "/") {
		InstallImageRepo = InstallImageRepo[:len(InstallImageRepo)-1]
	}
}

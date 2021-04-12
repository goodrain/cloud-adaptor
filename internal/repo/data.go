package repo

import (
	"github.com/google/wire"
	"goodrain.com/cloud-adaptor/internal/repo/appstore"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewCloudAccessKeyRepo,
	NewCreateKubernetesTaskRepo,
	NewInitRainbondRegionTaskRepo,
	NewUpdateKubernetesTaskRepo,
	NewTaskEventRepo,
	NewRainbondClusterConfigRepo,
	NewAppStoreRepo,
	appstore.NewStorer,
)

package repo

import "github.com/google/wire"

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewCloudAccessKeyRepo,
	NewCreateKubernetesTaskRepo,
	NewInitRainbondRegionTaskRepo,
	NewUpdateKubernetesTaskRepo,
	NewTaskEventRepo,
	NewRainbondClusterConfigRepo,
	NewAppStoreRepo,
)

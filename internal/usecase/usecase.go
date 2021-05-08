package usecase

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewClusterUsecase,
	NewAppStoreUsecase,
	NewAppTemplate,
)

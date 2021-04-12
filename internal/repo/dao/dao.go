package dao

import (
	"github.com/google/wire"
)

// ProviderSet is dao providers.
var ProviderSet = wire.NewSet(
	NewAppStoreDao,
)

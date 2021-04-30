package v1

import "github.com/helm/helm/pkg/repo"

// CreateAppStoreReq -
type CreateAppStoreReq struct {
	// The name of app store.
	Name     string `json:"name" binding:"required,appStoreName"`
	// The url of app store.
	URL      string `json:"url" binding:"required"`
	// The branch of app store, which category is git repo.
	Branch   string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// UpdateAppStoreReq -
type UpdateAppStoreReq struct {
	// The url of app store.
	URL      string `json:"url" binding:"required"`
	// The branch of app store, which category is git repo.
	Branch   string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// AppStore -
type AppStore struct {
	// The enterprise id.
	EID string `json:"eid"`
	// The name of app store.
	Name string `json:"name"`
	// The url of app store.
	URL string `json:"url"`
	// The branch of app store, which category is git repo.
	Branch string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// AppTemplate -
type AppTemplate struct {
	// The name of app template.
	Name     string          `json:"name"`
	// A list of app template versions.
	Versions []*repo.ChartVersion `json:"versions"`
}

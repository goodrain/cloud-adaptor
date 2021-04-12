package v1

import "github.com/helm/helm/pkg/repo"

// CreateAppStoreReq -
type CreateAppStoreReq struct {
	// TODO: 使用 HELM 的格式来校验 name
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Branch   string `json:"branch"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateAppStoreReq -
type UpdateAppStoreReq struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AppStore -
type AppStore struct {
	EID        string `json:"eid"`
	AppStoreID string `json:"appStoreID"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Branch     string `json:"branch"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// AppTemplate -
type AppTemplate struct {
	Name     string               `json:"name"`
	Versions []*repo.ChartVersion `json:"versions"`
}

package v1

// CreateAppStoreReq -
type CreateAppStoreReq struct {
	// TODO: 使用 HELM 的格式来校验 name
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Branch   string `json:"branch"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateAppStoreResp -
type CreateAppStoreResp struct {
	EID      string `json:"eid"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Username string `json:"username"`
	Password string `json:"password"`
}

package bcode

// appstore 8000 ~ 8999
var (
	ErrAppStoreNotFound     = newByMessage(404, 8000, "app store not found")
	ErrAppStoreNameConflict = newByMessage(409, 8001, "app store name conflict")
)
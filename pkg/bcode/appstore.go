package bcode

// appstore 8000 ~ 8999
var (
	ErrAppStoreNotFound        = newByMessage(404, 8000, "app store not found")
	ErrAppStoreNameConflict    = newByMessage(409, 8001, "app store name conflict")
	ErrAppStoreUnavailable     = newByMessage(400, 8002, "app store unavailable")
	ErrAppTemplateNotFound     = newByMessage(404, 8003, "app template not found")
	ErrTemplateVersionNotFound = newByMessage(404, 8004, "template version not found")
)

package uuidutil

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

// NewUUID new uuid
func NewUUID() string {
	uid := uuid.NewV4().String()
	return strings.Replace(uid, "-", "", -1)
}

package uuidutil

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

// NewUUID new uuid
func NewUUID() string {
	uid, _ := uuid.NewV4()
	return strings.Replace(uid.String(), "-", "", -1)
}

package dao

import "strings"

func isDuplicateEntry(err error) bool {
	return strings.HasPrefix(err.Error(), "UNIQUE constraint failed") ||
		strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry")
}

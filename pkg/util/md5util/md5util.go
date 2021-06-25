package md5util

import (
	"crypto/md5"
	"fmt"
)

//Md5Crypt -
func Md5Crypt(encryptStr, salt string) (CryptStr string) {
	if salt == "" {
		salt = "goodrain"
	}
	str := fmt.Sprintf("%s%s", encryptStr, salt)
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

package util

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

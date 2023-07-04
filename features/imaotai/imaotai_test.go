package imaotai

import (
	"encoding/base64"
	"testing"
)

func Test_version(t *testing.T) {
	t.Log(version())
}

func Test_getCode(t *testing.T) {
	decodeString, _ := base64.StdEncoding.DecodeString("eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9")
	t.Log(string(decodeString))
	decodeString, _ = base64.StdEncoding.DecodeString("eyJpc3MiOiJtdCIsImV4cCI6MTY5MTAzMTI4NCwidXNlcklkIjoxMDc3NjM3NzU5LCJkZXZpY2VJZCI6IjJGMjA3NUQwLUI2NkMtNDI4Ny1BOTAzLURCRkY2MzU4MzQyQSIsImlhdCI6MTY4ODQzOTI4NH0")
	t.Log(string(decodeString))
	decodeString, _ = base64.StdEncoding.DecodeString("ebjcK9dOimKFXXvcpUkBc_BkVM1YkWAimmc28hb9gQA")
	t.Log(string(decodeString))
}

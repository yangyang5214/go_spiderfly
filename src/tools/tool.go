package tools

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func Md5(data []byte) [16]byte {
	return md5.Sum(data)
}

func StringToByte(data []string) []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(data)
	return buf.Bytes()
}

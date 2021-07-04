package database

import (
	"reflect"
	"unsafe"
)

func S2B(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

package storage

import (
	"crypto/rand"
	"encoding/base64"
)

// map как хранилище
var MyMap = make(map[string]string)

type MemoryStorage map[string]string

type StorageServise struct {
	Storage map[string]string
}

func NewStorageServise() *StorageServise {
	return &StorageServise{MyMap}
}

// обработчи для короткой строки
func RandStringBytes(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

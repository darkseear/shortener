package storage

import "math/rand"

// map как хранилище
var MyMap = make(map[string]string)

type StorageServise struct {
	Storage map[string]string
}

func NewStorageServise() *StorageServise {
	return &StorageServise{MyMap}
}

//строка из которой будем брать символы для преобразования в короткую строку
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//обработчи для короткой строки
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

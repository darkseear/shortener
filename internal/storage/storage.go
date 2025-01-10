package storage

import "math/rand"

// map как хранилище
var MyMap = make(map[string]string)

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

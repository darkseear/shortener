package client

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	// "net/url"
	// "strconv"
)

func main() {
	endpoint := "http://localhost:8080/"

	// контейнер данных для запроса
	// data := url.Values{}

	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")

	// заполняем контейнер данными
	// data.Set("url", long)

	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(long))
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "text/plain")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	// и печатаем его
	fmt.Println(string(body))

	//гет запрос
	request2, err2 := http.NewRequest(http.MethodGet, string(body), nil)
	if err2 != nil {
		panic(err2)
	}

	response2, err3 := client.Do(request2)
	if err3 != nil {
		panic(err3)
	}

	defer response2.Body.Close()
	fmt.Println("Статус-код ", response2.Status)

}

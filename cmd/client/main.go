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
	request_2, err_2 := http.NewRequest(http.MethodGet, string(body), nil)
	if err_2 != nil {
		panic(err_2)
	}

	response_2, err_3 := client.Do(request_2)
	if err_3 != nil {
		panic(err_3)
	}

	defer response_2.Body.Close()
	fmt.Println("Статус-код ", response_2.Status)

}

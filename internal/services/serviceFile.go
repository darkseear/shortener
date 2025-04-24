package services

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
)

// Producer - структура для записи в файл.
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// Consumer - структура для чтения из файла.
type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// NewProducer - создает новый экземпляр Producer.
// Принимает имя файла в качестве аргумента и возвращает указатель на Producer и ошибку.
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// WriteMemoryFile - записывает объект MemoryFile в файл в формате JSON.
// Принимает указатель на MemoryFile в качестве аргумента и возвращает ошибку.
func (p *Producer) WriteMemoryFile(memoryFile *models.MemoryFile) error {
	return p.encoder.Encode(&memoryFile)
}

// Close - закрывает файл, связанный с Producer.
// Возвращает ошибку, если закрытие файла не удалось.
func (p *Producer) Close() error {
	return p.file.Close()
}

// NewConsumer - создает новый экземпляр Consumer.
// Принимает имя файла в качестве аргумента и возвращает указатель на Consumer и ошибку.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// ReadMemoryFile - читает объект MemoryFile из файла в формате JSON.
// Возвращает указатель на MemoryFile и ошибку.
func (c *Consumer) ReadMemoryFile() (*models.MemoryFile, error) {
	event := &models.MemoryFile{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}

	return event, nil
}

// ReadMemoryFileAll - читает все объекты MemoryFile из файла в формате JSON.
// Возвращает карту, где ключ - короткий URL, а значение - длинный URL, и ошибку.
func (c *Consumer) ReadMemoryFileAll() (map[string]string, error) {
	result := map[string]string{}
	line := &models.MemoryFile{}

	for {
		err := c.decoder.Decode(&line)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		result[line.ShortURL] = line.LongURL
	}

	return result, nil
}

// Close - закрывает файл, связанный с Consumer.
func (c *Consumer) Close() error {
	return c.file.Close()
}

// MemoryFileSave - сохраняет данные из MemoryStorage в файл.
// Принимает имя файла и указатель на MemoryStorage в качестве аргументов и возвращает ошибку.
func MemoryFileSave(filename string, m *MemoryStorage) error {

	Producer, err := NewProducer(filename)
	if err != nil {
		logger.Log.Error("no file")
		return err
	}
	defer Producer.Close()

	Consumer, err := NewConsumer(filename)
	if err != nil {
		logger.Log.Error("no file")
		return err
	}

	defer Consumer.Close()

	readMemoryFile, err := Consumer.ReadMemoryFileAll()
	if err != nil {
		logger.Log.Error("no read file")
		return err
	}

	for key, item := range readMemoryFile {
		m.Memory[key] = item
	}

	return nil
}

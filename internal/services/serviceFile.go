package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

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

func (p *Producer) WriteMemoryFile(memoryFile *models.MemoryFile) error {
	return p.encoder.Encode(&memoryFile)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

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

func (c *Consumer) ReadMemoryFile() (*models.MemoryFile, error) {
	event := &models.MemoryFile{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}

	return event, nil
}

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

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (m *LocalMemory) MemoryFileSave(filename string) (*LocalMemory, error) {

	// m := NewMemory()

	Producer, err := NewProducer(filename)
	if err != nil {
		logger.Log.Error("no file")
		return nil, err
	}
	defer Producer.Close()

	Consumer, err := NewConsumer(filename)
	if err != nil {
		logger.Log.Error("no file")
		return nil, err
	}

	defer Consumer.Close()

	readMemoryFile, err := Consumer.ReadMemoryFileAll()
	if err != nil {
		logger.Log.Error("no read file")
		return nil, err
	}

	for key, item := range readMemoryFile {
		m.localMemory.Memory[key] = item
	}

	fmt.Println(readMemoryFile)
	fmt.Println(m.localMemory.Memory)
	return m, nil
}

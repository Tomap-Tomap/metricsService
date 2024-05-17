// Package file provides structures for working with files.
//
// Deprecated
package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Producer struct {
	file    *os.File
	Encoder *json.Encoder
	m       sync.Mutex
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		Encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) ClearFile() error {
	return p.file.Truncate(0)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func (p *Producer) WriteInFile(value any) error {
	p.m.Lock()
	defer p.m.Unlock()

	if err := p.Encoder.Encode(value); err != nil {
		return fmt.Errorf("encode value in file: %w", err)
	}

	return nil
}

type Consumer struct {
	file    *os.File
	Decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		Decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

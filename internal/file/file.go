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

// Producer structure for working with writing data to a file
type Producer struct {
	file    *os.File
	Encoder *json.Encoder
	m       sync.Mutex
}

// NewProducer create Producer
func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		Encoder: json.NewEncoder(file),
	}, nil
}

// ClearFile clears file
func (p *Producer) ClearFile() error {
	err := p.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = p.file.Seek(0, 0)
	return err
}

// Close closes Producer
func (p *Producer) Close() error {
	return p.file.Close()
}

// WriteInFile writes value in file
func (p *Producer) WriteInFile(value any) error {
	p.m.Lock()
	defer p.m.Unlock()

	if err := p.Encoder.Encode(value); err != nil {
		return fmt.Errorf("encode value in file: %w", err)
	}

	return nil
}

// Consumer a structure for working with reading data from a file
type Consumer struct {
	file    *os.File
	Decoder *json.Decoder
}

// NewConsumer create Consumer
func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		Decoder: json.NewDecoder(file),
	}, nil
}

// Close closes Consumer
func (c *Consumer) Close() error {
	return c.file.Close()
}

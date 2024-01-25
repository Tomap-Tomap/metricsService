package file

import (
	"encoding/json"
	"os"
)

type Producer struct {
	file    *os.File
	Encoder *json.Encoder
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

func (p *Producer) Seek() error {
	_, err := p.file.Seek(0, 0)
	return err
}

func (p *Producer) Close() error {
	return p.file.Close()
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

package repository

import (
	"context"
	"encoding/json"
	"os"

	"github.com/coolycow/shortener/internal/model"
	"github.com/google/uuid"
)

type JSONStorage struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func (s *JSONStorage) Close() error {
	return s.file.Close()
}

func (s *JSONStorage) Load(repo URLRepository) error {
	for s.decoder.More() {
		var url model.ShortURL

		if err := s.decoder.Decode(&url); err != nil {
			return err
		}

		if _, _, err := repo.AddURL(context.Background(), uuid.New().String(), url.OriginalURL, url.Key); err != nil {
			return err
		}
	}

	return nil
}

func (s *JSONStorage) Write(url model.ShortURL) error {
	err := s.encoder.Encode(url)

	if err != nil {
		return err
	}

	return nil
}

func NewJSONStorage(filename string) (*JSONStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	return &JSONStorage{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}, nil
}

package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
)

const (
	fileFlags = syscall.O_RDWR | syscall.O_CREAT | syscall.O_APPEND | syscall.O_SYNC
	filePerms = 0644
)

var (
	ErrEmptyFile = errors.New("empty file")
)

type fileStorage struct {
	file *os.File
}

func New(cfg config.BakConfig) (*fileStorage, error) {
	const fn = "fileStorage.New"

	file, err := os.OpenFile(cfg.Path, fileFlags, filePerms)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	//Являет ли такое создание гарантией того, что в file не будет nil?
	return &fileStorage{
		file: file,
	}, nil
}

func (f *fileStorage) Close() error {
	return f.file.Close()
}

func (f *fileStorage) Save(data []models.Metrics) error {
	const fn = "fileStorage.Save"

	buf := bufio.NewWriter(f.file)
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	return buf.Flush()
}

func (f *fileStorage) Restore() ([]models.Metrics, error) {
	const fn = "fileStorage.Restore"

	strData, err := f.getLastJSON()
	if err != nil {
		if errors.Is(err, ErrEmptyFile) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	var data []models.Metrics

	if err := json.Unmarshal([]byte(strData), &data); err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	return data, nil
}

func (f *fileStorage) readAllJSON() ([]string, error) {
	const fn = "fileStorage.readAllJSON"

	if _, err := f.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	data, err := io.ReadAll(f.file)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	content := string(data)

	jsons := strings.Split(content, "\n")

	if len(jsons) > 0 && jsons[len(jsons)-1] == "" {
		jsons = jsons[:len(jsons)-1]

	}

	return jsons, nil
}

func (f *fileStorage) getLastJSON() (string, error) {
	const fn = "fileStorage.getLastJSON"

	lines, err := f.readAllJSON()
	if err != nil {
		return "", fmt.Errorf("%s: %v", fn, err)
	}

	if len(lines) == 0 {
		return "", ErrEmptyFile
	}

	return lines[len(lines)-1], nil
}

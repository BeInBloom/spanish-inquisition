package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
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

func New(path string) (*fileStorage, error) {
	const fn = "fileStorage.New"

	file, err := os.OpenFile(path, fileFlags, filePerms)
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

func (f *fileStorage) Save(data []ptypes.Metrics) error {
	const fn = "fileStorage.Save"

	strData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	//если в json будет строка c \n то все сломается, я знаю
	//причем шанс на ее появляение не нулевой
	strData = append(strData, '\n')

	_, err = f.file.Write(strData)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	return nil
}

func (f *fileStorage) Restore() ([]ptypes.Metrics, error) {
	const fn = "fileStorage.Restore"

	strData, err := f.getLastJSON()
	if err != nil {
		if errors.Is(err, ErrEmptyFile) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	var data []ptypes.Metrics

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

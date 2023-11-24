package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

const (
	defaultFileExt = "yaml"
)

var (
	errNoServiceConfig       = fmt.Errorf("the service's configuration must be adjusted")
	errMustBePointer         = fmt.Errorf("configPtr arg must be pointer")
	errMustBePointerToStruct = fmt.Errorf("configPtr arg must be pointer to struct")
	errFailedToLoadConfig    = fmt.Errorf("cannot load from config")
	errEmptyPath             = fmt.Errorf("emptyPath")
	errFileNotExists         = fmt.Errorf("file not exists")
	errMustBeFile            = fmt.Errorf("must be file")
	errWrongExtension        = fmt.Errorf("config file must be yaml")
)

func LoadConfig(configStruct interface{}, defCfgFile string) error {
	if configStruct == nil {
		return errNoServiceConfig
	}

	return configure(configStruct, defCfgFile)
}

func configure(configPtr interface{}, configFilePath string) error {
	{ // target type validation
		t := reflect.TypeOf(configPtr)
		if t.Kind() != reflect.Ptr {
			return errMustBePointer
		}
		if t.Elem().Kind() != reflect.Struct {
			return errMustBePointerToStruct
		}
	}

	loadedFromFile, err := loadFromFile(configPtr, configFilePath)
	if err != nil {
		return err
	}

	if !loadedFromFile {
		return errFailedToLoadConfig
	}

	return nil
}

func loadFromFile(
	target interface{},
	configPath string,
) (bool, error) {
	if err := fileExists(configPath); err != nil {
		return false, err
	}

	if err := checkFileExt(configPath); err != nil {
		return false, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return false, err
	}
	if err = yaml.NewDecoder(file).Decode(target); err != nil {
		return false, err
	}

	return true, nil
}

func fileExists(path string) (err error) {
	if path == "" {
		return errEmptyPath
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", errFileNotExists, absPath)
		}
		return
	}
	if info.IsDir() {
		return fmt.Errorf("%w: %s", errMustBeFile, absPath)
	}
	return
}

func checkFileExt(path string) error {
	if path == "" {
		return errEmptyPath
	}
	ext := filepath.Ext(path)
	if len(ext) > 1 {
		ext = ext[1:]
	}

	if ext != defaultFileExt {
		return errWrongExtension
	}

	return nil
}

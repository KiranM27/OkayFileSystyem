package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateFolder(path string) {
	if pathExists, _ := PathExists(path); !pathExists {
		err := os.Mkdir(path, 0777)
		if err != nil {
			fmt.Println("Error while creating ", path, " : ", err)
		}
	}
}

func CreateFile(path string) {
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	defer fh.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func ReadFile(path string) string {
	if pathExists, _ := PathExists(path); pathExists {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("Error while reading file of path - ", path, " : ", err)
		}
		return string(content)
	}
	return NULL
}

// Helper to remove extensions
func RemoveExtension(fpath string) string {
	ext := filepath.Ext(fpath)
	return strings.TrimSuffix(fpath, ext)
}

func AppendToFile(fpath string, content string) {
	content = content + "\n"
	f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error while opening file ", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		fmt.Println("Error while appending to file ", err)
	}
}

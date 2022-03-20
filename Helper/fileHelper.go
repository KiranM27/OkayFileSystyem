package helper

import (
	"os"
	"fmt"
)

func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
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
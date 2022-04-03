package client

import (
	"os"
	"fmt"
	"log"
	"math"
	"time"
	"bytes"
	"path/filepath"
	"strconv"
	"io/ioutil"
	"strings"
	"encoding/json"
	"net/http"
	structs "oks/Structs"
	helper "oks/Helper"

)

// Destination Port (increased Pointer) --> message.Port[message.Pointer], increase Pointer before sending
func SendTimerMessage(message structs.Message) bool {
	portNo := message.Ports[message.Pointer]
	request_url := helper.BASE_URL + ":" + strconv.Itoa(portNo) + "/message"
	messageJSON, _ := json.Marshal(message)


	_client := &http.Client{
			Timeout: time.Minute,
	}
	resp, err := _client.Post(request_url, "application/json", bytes.NewBuffer(messageJSON))

	//Handle Error
	if err != nil {
		log.Fatalf("SendMessage: An Error Occured - %s", err)
	}

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	log.Println(string(body))
	return true
}


// Get size of file trying to write 
func getFileSize(filename string) (int64) {
	
	// Get relative path of the text file
	// First arg is the main directory, second arg is where file is stored
	fmt.Println(os.Getwd())
	rel, err := filepath.Rel(helper.ROOT_DIR + helper.TEST_DIR, helper.ROOT_DIR + helper.TEST_DATA_DIR + "/" + filename)
    if err != nil {
        panic(err)
    }
    fmt.Println("getFileSize path ", rel) // debug

	file, _ := os.Open(rel)
	fi, err := file.Stat()
	if err != nil {
	// Could not obtain stat, handle error
	}
	fmt.Printf("%s is %d bytes long\n", filename, fi.Size()) // debug
	return fi.Size()

}

// Used to split files that are larger than 2.5kb
func splitFile(oldFilename string)(uint64){
		
		// Relative file path
		rel, err := filepath.Rel(helper.ROOT_DIR + helper.TEST_DIR, helper.ROOT_DIR + helper.TEST_DATA_DIR + "/" + oldFilename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Open the file
        file, err := os.Open(rel)
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
		defer file.Close()

		// Get file data
		fileInfo, _ := file.Stat()
		var fileSize int64 = fileInfo.Size()
		const fileChunk = 2500 // 2.5 KB

		// calculate total number of parts the file will be chunked into
		totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
		fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
		for i := uint64(0); i < totalPartsNum; i++ {

			partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
			partBuffer := make([]byte, partSize)
			file.Read(partBuffer)

			// write to disk
			rel, err := filepath.Rel(helper.ROOT_DIR + helper.TEST_DIR, helper.ROOT_DIR + helper.TEST_DATA_DIR + "/" + oldFilename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			rel = removeExtension(rel)
			fileName := rel + strconv.FormatUint(i, 10) + ".txt"
			_, err = os.Create(fileName)
			if err != nil {
					fmt.Println(err)
					os.Exit(1)
			}
			// write/save buffer to disk
			ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)
			fmt.Println("Split to : ", fileName)
		}
		return totalPartsNum
}

// Helper to remove extensions [Done]
func removeExtension(fpath string) string {
	ext := filepath.Ext(fpath)
	return strings.TrimSuffix(fpath, ext)
}

// Helper to read file
func readFile(filename string) string {
	// Read the file
	rel, err := filepath.Rel(helper.ROOT_DIR + helper.TEST_DIR, helper.ROOT_DIR + helper.TEST_DATA_DIR + "/" + filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	text, err := ioutil.ReadFile(rel)
    if err != nil {
        fmt.Print(err)
		os.Exit(1)
    }
	return string(text)
}
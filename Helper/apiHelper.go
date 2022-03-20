package helper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	//structs "gfs.com/master/structs"
)


// Destination Port (increased Pointer) --> message.Port[message.Pointer], increase Pointer before sending
func SendMessage(portNo int, message structs.Message) { // V2 takes in a Message object directly.
	request_url := BASE_URL + ":" + strconv.Itoa(portNo) + "/message"
	messageJSON, _ := json.Marshal(message)
	response, err := http.Post(request_url, "application/json", bytes.NewBuffer(messageJSON))

	//Handle Error
	if err != nil {
		log.Fatalf("SendMessage: An Error Occured - %v", err)
	}

	defer response.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(body))
}

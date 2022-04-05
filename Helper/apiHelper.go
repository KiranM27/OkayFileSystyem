package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	structs "oks/Structs"
	"strconv"
)

// Destination Port (increased Pointer) --> message.Port[message.Pointer], increase Pointer before sending
func SendMessage(message structs.Message) { // V2 takes in a Message object directly.
	portNo := message.Ports[message.Pointer]
	request_url := BASE_URL + ":" + strconv.Itoa(portNo) + "/message"
	messageJSON, _ := json.Marshal(message)
	response, err := http.Post(request_url, "application/json", bytes.NewBuffer(messageJSON))

	//Handle Error (do not log.Fatal so that it doesnt exit)
	if err != nil {
		log.Printf("SendMessage: An Error Occured - %v", err)
	}else{
		defer response.Body.Close()
		//Read the response body (do not log.Fatal so that it doesnt exit)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("SendMessage: An Error Occured at Body - %v",err)
		}
		log.Println(string(body))
	}

}

func SendRep(repMsg structs.RepMsg){
	portNo := repMsg.TargetCS
	request_url := BASE_URL + ":" + strconv.Itoa(portNo) + "/replicate"
	messageJSON, _ := json.Marshal(repMsg)
	response, err := http.Post(request_url, "application/json", bytes.NewBuffer(messageJSON))

	//Handle Error (do not log.Fatal so that it doesnt exit)
	if err != nil {
		fmt.Printf("SendRep: An Error Occured - %v\n", err)
	}else{
		defer response.Body.Close()
		//Read the response body (do not log.Fatal so that it doesnt exit)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("SendRep: An Error Occured at Body - %v\n",err)
		}
		fmt.Println(string(body))
	}
}

package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"io/ioutil"

	quic "github.com/quic-go/quic-go"
)

type ClientConfig struct {
	Host string `json:"host"`
}

func loadConfig(path string) (ClientConfig, error) {
	var config ClientConfig
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func sendMessage(stream quic.Stream, msg Message) {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, uint32(len(body)))

	_, err = stream.Write(lengthPrefix)
	if err != nil {
		log.Println("Error writing length prefix:", err)
		return
	}

	_, err = stream.Write(body)
	if err != nil {
		log.Println("Error writing message body:", err)
	}
}

func readMessage(stream quic.Stream) (Message, error) {
	var msg Message

	header := make([]byte, 4)
	_, err := io.ReadFull(stream, header)
	if err != nil {
		if err == io.EOF {
			return msg, err
		}
		return msg, errors.New("failed to read message length: " + err.Error())
	}
	length := binary.BigEndian.Uint32(header)

	body := make([]byte, length)
	_, err = io.ReadFull(stream, body)
	if err != nil {
		return msg, errors.New("failed to read message body: " + err.Error())
	}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		return msg, errors.New("failed to decode JSON message: " + err.Error())
	}

	return msg, nil
}
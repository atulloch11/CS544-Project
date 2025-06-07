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

// ClientConfig holds configuration values loaded from a JSON config file.
// currently it includes only the server hostname or IP address.
type ClientConfig struct {
	Host string `json:"host"`
}

// method reads and parses a JSON config file from the specified path.
// it returns a ClientConfig object or an error if loading or decoding fails.
func loadConfig(path string) (ClientConfig, error) {
	var config ClientConfig
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

// method encodes a Message struct into JSON, prepends it with a 4-byte length
// prefix, and sends it over the provided QUIC stream.
// this ensures the receiver knows exactly how many bytes to read for the full message.
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

// method reads a full message from the given QUIC stream by first reading a 4-byte
// length prefix, then reading and decoding the JSON payload of that length.
// it returns a decoded Message struct or an error if reading or decoding fails.
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
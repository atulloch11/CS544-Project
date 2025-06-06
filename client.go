package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	quic "github.com/quic-go/quic-go"
)

var currentState ProtocolState = StateStart

// Method to run the client and connect to server
func runClient(host string) {
	reader := bufio.NewReader(os.Stdin)
	addr := fmt.Sprintf("%s:%d", host, ServerPort)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"qtgp-demo"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConfig, nil)
	if err != nil {
		log.Fatal("[CLIENT] ❌ Failed to connect:", err)
	}
	defer conn.CloseWithError(0, "client shutdown")
	fmt.Printf("[CLIENT] ✅ Connected to server at %s\n", addr)

	// print out user readable game menu
	for {
		fmt.Println("\n🎮 GAME MENU")
		fmt.Println("1. Start Game")
		fmt.Println("2. Make Move")
		fmt.Println("3. Resync Game State")
		fmt.Println("4. Quit")
		fmt.Print("Choose an option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			sendJoinGame(conn)
		case "2":
			sendStateUpdate(conn)
		case "3":
			sendResyncRequest(conn)
		case "4":
			fmt.Println("[CLIENT] 👋 Exiting game. Goodbye!")
			return
		default:
			fmt.Println("[CLIENT] ⚠️ Invalid choice. Please try again.")
		}
	}
}

func sendJoinGame(conn quic.Connection) {
	if currentState != StateStart {
		log.Printf("[CLIENT] ⚠️ Cannot start game in current state: %v\n", currentState)
		return
	}
	transitionTo(&currentState, StateWaitingForJoin)

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to open stream:", err)
		return
	}
	defer stream.Close()

	msg := Message{
		ProtocolVersion: 1,
		Type:            "JOIN_GAME_REQUEST",
		PlayerID:        "ashley123",
		GameID:          "civilization_game",
		TurnOptions:     0b00000001,
	}
	sendMessage(stream, msg)
	transitionTo(&currentState, StateJoining)

	ack, err := readMessage(stream)
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to receive GAME_SETUP_ACK:", err)
		return
	}
	if ack.Type == "GAME_SETUP_ACK" {
		fmt.Println("[CLIENT] ✅ Game Setup Acknowledged")
		fmt.Printf("         Status: Success - Join Accepted")
		fmt.Printf("         Agreed Options: Turn Timeouts Off")
		transitionTo(&currentState, StateInGame)
	}
}

func sendStateUpdate(conn quic.Connection) {
	if currentState != StateInGame {
		log.Printf("[CLIENT] ⚠️ Cannot make move in current state: %v\n", currentState)
		return
	}
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to open stream:", err)
		return
	}
	defer stream.Close()

	msg := Message{
		ProtocolVersion: 1,
		Type:            "STATE_UPDATE",
		GameState:       "TURN_1:PLAYER_ASHLEY_MOVE",
	}
	sendMessage(stream, msg)

	ack, err := readMessage(stream)
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to receive STATE_ACK:", err)
		return
	}
	if ack.Type == "STATE_ACK" {
		fmt.Println("[CLIENT] ✅ Move acknowledged by server.")
	}
}

func sendResyncRequest(conn quic.Connection) {
	if currentState != StateInGame {
		log.Printf("[CLIENT] ⚠️ Cannot request resync in current state: %v\n", currentState)
		return
	}
	transitionTo(&currentState, StateResyncing)

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to open stream:", err)
		return
	}
	defer stream.Close()

	msg := Message{
		ProtocolVersion: 1,
		Type:            "STATE_RESYNC_REQUEST",
	}
	sendMessage(stream, msg)

	ack, err := readMessage(stream)
	if err != nil {
		log.Println("[CLIENT] ❌ Failed to receive STATE_ACK:", err)
		return
	}
	if ack.Type == "STATE_ACK" {
		fmt.Println("[CLIENT] 🔄 Game state successfully resynced.")
		transitionTo(&currentState, StateInGame)
	}
}

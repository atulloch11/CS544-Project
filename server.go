package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	quic "github.com/quic-go/quic-go"
)

// method to start a QUIC sever to listen on the user provided host and hardcoded port
// continously accepts incoming connections and spawns goroutines to handle each one
func runServer(host string) {
	addr := fmt.Sprintf("%s:%d", host, ServerPort)
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		log.Fatal("[SERVER] Failed to start server:", err)
	}
	log.Printf("[SERVER] Listening on %s\n", addr)

	for {
		log.Println("[SERVER] Waiting for client connection...")
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Println("[SERVER] Accept error:", err)
			continue
		}
		log.Println("[SERVER] Accepted new connection!")
		go handleClient(conn)
	}
}

// manages a single QUIC connection from a client
// each client connection maintains its own protocol state for DFA logic
// each separate stream opened by client, creats a goroutine
func handleClient(conn quic.Connection) {
	defer conn.CloseWithError(0, "client handler done")
	log.Println("[SERVER] Handling new client...")

	// Each client connection gets its own state
	clientState := StateStart

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Println("[SERVER] Stream accept error or connection closed:", err)
			return
		}
		go handleStream(stream, &clientState)
	}
}

// read and process a message on an incoming stream
// validates the client's current protocol state, performs transitions,
// sends response message based on message type to client
func handleStream(stream quic.Stream, clientState *ProtocolState) {
	msg, err := readMessage(stream)
	if err != nil {
		if err == io.EOF || strings.Contains(err.Error(), "application error") {
			log.Println("[SERVER] Client closed stream.")
			return
		}
		log.Println("[SERVER] Failed to read message:", err)
		return
	}

	log.Printf("[SERVER] Received: %s | Player: %s | Game: %s | State: %s\n",
		msg.Type, msg.PlayerID, msg.GameID, msg.GameState)

	switch msg.Type {

	// handle client request to join a game
	case "JOIN_GAME_REQUEST":
		if *clientState == StateStart {
			transitionTo(clientState, StateWaitingForJoin)
		}
	
		if *clientState != StateWaitingForJoin {
			log.Printf("[SERVER] Invalid state (%v) for JOIN_GAME_REQUEST\n", *clientState)
			return
		}
	
		transitionTo(clientState, StateJoining)
	
		ack := Message{
			ProtocolVersion: msg.ProtocolVersion,
			Type:            "GAME_SETUP_ACK",
			Status:          0x00,
			AgreedOptions:   msg.TurnOptions,
		}
		sendMessage(stream, ack)
		transitionTo(clientState, StateInGame)

	// handle client sending a game state update (i.e. moving in game)
	case "STATE_UPDATE":
		if *clientState != StateInGame {
			log.Printf("[SERVER] Invalid state (%v) for STATE_UPDATE\n", *clientState)
			return
		}
		log.Printf("[SERVER] Game state updated: %s\n", msg.GameState)

		sendMessage(stream, Message{
			ProtocolVersion: msg.ProtocolVersion,
			Type:            "STATE_ACK",
		})

	// handle client requesting a resync to game state
	case "STATE_RESYNC_REQUEST":
		if *clientState != StateInGame {
			log.Printf("[SERVER] Invalid state (%v) for STATE_RESYNC_REQUEST\n", *clientState)
			return
		}
		transitionTo(clientState, StateResyncing)

		log.Println("[SERVER] Resync requested.")
		sendMessage(stream, Message{
			ProtocolVersion: msg.ProtocolVersion,
			Type:            "STATE_ACK",
		})
		transitionTo(clientState, StateInGame)

	default:
		log.Printf("[SERVER] Unknown message type: %s\n", msg.Type)
	}
}

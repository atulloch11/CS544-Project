package main

// message defines the structure of a protocol message exchanged between the client and server.
// fields are tagged for JSON serialization to allow consistent encoding/decoding across the QUIC stream.
type Message struct {
	ProtocolVersion int    `json:"protocol_version"`
	Type            string `json:"type"`
	PlayerID    string `json:"player_id,omitempty"`
	GameID      string `json:"game_id,omitempty"`
	TurnOptions uint8  `json:"turn_options,omitempty"`
	Status        uint8 `json:"status,omitempty"`
	AgreedOptions uint8 `json:"agreed_options,omitempty"`
	GameState       string `json:"game_state,omitempty"`
}

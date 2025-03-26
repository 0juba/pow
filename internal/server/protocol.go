package server

import (
	"encoding/json"
	"fmt"
	"net"
)

type MessageType string

const (
	RequestChallenge MessageType = "request_challenge"
	SubmitSolution   MessageType = "submit_solution"
	RequestQuote     MessageType = "request_quote"
	Response         MessageType = "response"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	Stamp       string `json:"stamp"`
}

type SolutionRequest struct {
	ChallengeID string `json:"challenge_id"`
	Solution    uint64 `json:"solution"`
}

type ResponseMessage struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type QuoteRequest struct {
	SessionID string `json:"session_id"`
}

func ReadMessage(conn net.Conn) (*Message, error) {
	var msg Message
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		return nil, fmt.Errorf("read message: %w", err)
	}
	return &msg, nil
}

func WriteMessage(conn net.Conn, msg *Message) error {
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

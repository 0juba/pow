package client

import (
	"encoding/json"
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

type QuoteRequest struct {
	SessionID string `json:"session_id"`
}

type ResponseMessage struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

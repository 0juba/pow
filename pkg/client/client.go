package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	conn net.Conn
	addr string
}

type Config struct {
	Addr           string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		ConnectTimeout: 5 * time.Second,
		RequestTimeout: 10 * time.Second,
	}
}

func New(cfg Config) *Client {
	return &Client{
		addr: cfg.Addr,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", c.addr, DefaultConfig().ConnectTimeout)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) send(msg interface{}) error {
	return json.NewEncoder(c.conn).Encode(msg)
}

func (c *Client) receive(msg interface{}) error {
	return json.NewDecoder(c.conn).Decode(msg)
}

func (c *Client) RequestChallenge() (*ChallengeResponse, error) {
	msg := Message{
		Type: RequestChallenge,
	}
	if err := c.send(msg); err != nil {
		return nil, fmt.Errorf("send challenge request: %w", err)
	}

	var resp Message
	if err := c.receive(&resp); err != nil {
		return nil, fmt.Errorf("receive challenge response: %w", err)
	}

	var respMsg ResponseMessage
	if err := json.Unmarshal(resp.Payload, &respMsg); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if respMsg.Error != "" {
		return nil, fmt.Errorf("server error: %s", respMsg.Error)
	}

	var challenge ChallengeResponse
	if err := json.Unmarshal(resp.Payload, &challenge); err != nil {
		return nil, fmt.Errorf("unmarshal challenge: %w", err)
	}

	return &challenge, nil
}

func (c *Client) SubmitSolution(challengeID string, solution uint64) (string, error) {
	req := SolutionRequest{
		ChallengeID: challengeID,
		Solution:    solution,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal solution request: %w", err)
	}

	msg := Message{
		Type:    SubmitSolution,
		Payload: reqData,
	}
	if err := c.send(msg); err != nil {
		return "", fmt.Errorf("send solution: %w", err)
	}

	var resp Message
	if err := c.receive(&resp); err != nil {
		return "", fmt.Errorf("receive solution response: %w", err)
	}

	var respMsg ResponseMessage
	if err := json.Unmarshal(resp.Payload, &respMsg); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if respMsg.Error != "" {
		return "", fmt.Errorf("server error: %s", respMsg.Error)
	}

	sessionID, ok := respMsg.Data.(map[string]interface{})["session_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid session_id in response")
	}

	return sessionID, nil
}

func (c *Client) GetQuote(sessionID string) (string, error) {
	req := QuoteRequest{
		SessionID: sessionID,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal quote request: %w", err)
	}

	msg := Message{
		Type:    RequestQuote,
		Payload: reqData,
	}
	if err := c.send(msg); err != nil {
		return "", fmt.Errorf("send quote request: %w", err)
	}

	var resp Message
	if err := c.receive(&resp); err != nil {
		return "", fmt.Errorf("receive quote response: %w", err)
	}

	var respMsg ResponseMessage
	if err := json.Unmarshal(resp.Payload, &respMsg); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if respMsg.Error != "" {
		return "", fmt.Errorf("server error: %s", respMsg.Error)
	}

	quote, ok := respMsg.Data.(map[string]interface{})["quote"].(string)
	if !ok {
		return "", fmt.Errorf("invalid quote in response")
	}

	return quote, nil
}

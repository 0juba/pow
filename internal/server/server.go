package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/0juba/pow/internal/pow"
	"github.com/0juba/pow/internal/storage/memory"
)

type Server struct {
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	sessionStorage   *memory.Storage
	challengeStorage *memory.Storage
	quoteStorage     *memory.QuoteStorage
	pow              pow.POWService
	difficulty       uint32
}

func NewServer(
	ctx context.Context,
	sessionStorage *memory.Storage,
	challengeStorage *memory.Storage,
	quoteStorage *memory.QuoteStorage,
	pow pow.POWService,
) *Server {
	ctx, cancel := context.WithCancel(ctx)
	return &Server{
		ctx:              ctx,
		cancel:           cancel,
		sessionStorage:   sessionStorage,
		challengeStorage: challengeStorage,
		quoteStorage:     quoteStorage,
		pow:              pow,
	}
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	s.listener = listener

	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("stop server: %w", err)
		}
	}
	s.wg.Wait()
	return nil
}

func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if s.ctx.Err() != nil {
					return
				}
				continue
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		s.wg.Done()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msg, err := ReadMessage(conn)
			if err != nil {
				_ = s.sendError(conn, fmt.Errorf("read message: %w", err))
				return
			}

			if err := s.handleMessage(conn, msg); err != nil {
				_ = s.sendError(conn, err)
			}
		}
	}
}

func (s *Server) handleMessage(conn net.Conn, msg *Message) error {
	switch msg.Type {
	case RequestChallenge:
		return s.handleChallengeRequest(conn, msg)
	case SubmitSolution:
		return s.handleSolutionRequest(conn, msg)
	case RequestQuote:
		return s.handleQuoteRequest(conn, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (s *Server) handleChallengeRequest(conn net.Conn, msg *Message) error {
	challenge, err := s.pow.GenerateChallenge(conn.RemoteAddr().String())
	if err != nil {
		return fmt.Errorf("generate challenge: %w", err)
	}

	if err := s.challengeStorage.Store(s.ctx, &memory.Item{
		ID:        challenge.ID,
		ExpiresAt: &challenge.ExpiresAt,
		Value:     *challenge,
	}); err != nil {
		return fmt.Errorf("store challenge: %w", err)
	}

	resp := ChallengeResponse{
		ChallengeID: challenge.ID.String(),
		Stamp:       challenge.Ch.Stamp(0),
	}

	respPayload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal challenge response: %w", err)
	}

	msg = &Message{
		Type:    Response,
		Payload: respPayload,
	}

	if err := WriteMessage(conn, msg); err != nil {
		return fmt.Errorf("write challenge response: %w", err)
	}

	return nil
}

func (s *Server) handleSolutionRequest(conn net.Conn, msg *Message) error {
	var req SolutionRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("unmarshal solution request: %w", err)
	}

	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		return fmt.Errorf("parse challenge id: %w", err)
	}

	item, err := s.challengeStorage.Get(s.ctx, challengeID)
	if err != nil {
		return fmt.Errorf("get challenge: %w", err)
	}

	challenge, ok := item.Value.(pow.ChallengeItem)
	if !ok {
		return fmt.Errorf("item value type should be pow.ChallengeItem, got: %T", item.Value)
	}

	valid, err := s.pow.VerifySolution(pow.Solution{
		ChallengeItem: challenge,
		Nonce:         req.Solution,
	})
	if err != nil {
		return fmt.Errorf("verify solution: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid solution")
	}

	sessionExpiresAt := time.Now().UTC().Add(24 * time.Hour)
	session := &memory.Item{
		ID:        uuid.New(),
		ExpiresAt: &sessionExpiresAt,
	}

	if err := s.sessionStorage.Store(s.ctx, session); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	if err := s.challengeStorage.Delete(s.ctx, challengeID); err != nil {
		return fmt.Errorf("delete challenge: %w", err)
	}

	resp := ResponseMessage{
		Success: true,
		Data: map[string]string{
			"session_id": session.ID.String(),
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		},
	}

	respPayload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	msg = &Message{
		Type:    Response,
		Payload: respPayload,
	}

	if err := WriteMessage(conn, msg); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func (s *Server) handleQuoteRequest(conn net.Conn, msg *Message) error {
	var req QuoteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("unmarshal quote request: %w", err)
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return fmt.Errorf("parse session id: %w", err)
	}

	session, err := s.sessionStorage.Get(s.ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	if session == nil || (session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now().UTC())) {
		return fmt.Errorf("session expired")
	}

	quote, err := s.quoteStorage.GetRandom()
	if err != nil {
		return fmt.Errorf("get random quote: %w", err)
	}

	resp := ResponseMessage{
		Success: true,
		Data: map[string]string{
			"quote": quote,
		},
	}

	respPayload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	msg = &Message{
		Type:    Response,
		Payload: respPayload,
	}

	if err := WriteMessage(conn, msg); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func (s *Server) sendError(conn net.Conn, err error) error {
	resp := &Message{
		Type: Response,
		Payload: json.RawMessage(fmt.Sprintf(`{
			"success": false,
			"error": "%s"
		}`, err.Error())),
	}
	return WriteMessage(conn, resp)
}

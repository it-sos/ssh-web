package ws

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ssh-web/internal/auth"
	"github.com/ssh-web/internal/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Handler struct {
	sessionStore   *auth.SessionStore
	sshConfig      *ssh.ClientConfig
	mu             sync.Mutex
	activeSessions int
	maxSessions    int
}

func NewHandler(store *auth.SessionStore, sshCfg *ssh.ClientConfig) *Handler {
	return &Handler{
		sessionStore: store,
		sshConfig:    sshCfg,
		maxSessions:  10,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := h.sessionStore.GetSessionToken(r)
	if _, ok := h.sessionStore.ValidateSession(token); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.mu.Lock()
	if h.activeSessions >= h.maxSessions {
		h.mu.Unlock()
		http.Error(w, "Too many sessions", http.StatusTooManyRequests)
		return
	}
	h.activeSessions++
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.activeSessions--
		h.mu.Unlock()
	}()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	sshClient, sshSession, err := ssh.Connect(h.sshConfig)
	if err != nil {
		slog.Error("SSH connection failed", "error", err)
		sendError(conn, "SSH connection failed: "+err.Error())
		return
	}
	defer sshSession.Close()
	defer sshClient.Close()

	if err := sshSession.RequestPty("xterm-256color", 24, 80, nil); err != nil {
		slog.Error("PTY request failed", "error", err)
		sendError(conn, "PTY request failed")
		return
	}

	if err := sshSession.Shell(); err != nil {
		slog.Error("Shell start failed", "error", err)
		sendError(conn, "Shell start failed")
		return
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		slog.Error("Stdin pipe failed", "error", err)
		return
	}
	defer stdin.Close()

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		slog.Error("Stdout pipe failed", "error", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "data":
				if data, ok := msg.Payload.(string); ok {
					stdin.Write([]byte(data))
				}
			case "resize":
				if payload, ok := msg.Payload.(map[string]interface{}); ok {
					cols := int(payload["cols"].(float64))
					rows := int(payload["rows"].(float64))
					if cols >= 1 && cols <= 500 && rows >= 1 && rows <= 200 {
						sshSession.WindowChange(rows, cols)
					}
				}
			case "close":
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if err := conn.WriteJSON(Message{
					Type:    "data",
					Payload: string(buf[:n]),
				}); err != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					slog.Error("SSH read error", "error", err)
				}
				return
			}
		}
	}()

	wg.Wait()
}

func sendError(conn *websocket.Conn, msg string) {
	conn.WriteJSON(Message{
		Type:    "error",
		Payload: msg,
	})
}

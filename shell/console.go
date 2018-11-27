package shell

import (
	"bytes"
	"strings"
	"time"
)

var (
	ansiColors = map[string][]byte{
		"none":         []byte("\x1b[0m"),
		"black":        []byte("\x1b[0;30m"),
		"red":          []byte("\x1b[0;31m"),
		"green":        []byte("\x1b[0;32m"),
		"orange":       []byte("\x1b[0;33m"),
		"blue":         []byte("\x1b[0;34m"),
		"purple":       []byte("\x1b[0;35m"),
		"cyan":         []byte("\x1b[0;36m"),
		"light-gray":   []byte("\x1b[0;37m"),
		"dark-gray":    []byte("\x1b[1;30m"),
		"light-red":    []byte("\x1b[1;31m"),
		"light-green":  []byte("\x1b[1;32m"),
		"yellow":       []byte("\x1b[1;33m"),
		"light-blue":   []byte("\x1b[1;34m"),
		"light-purple": []byte("\x1b[1;35m"),
		"light-cyan":   []byte("\x1b[1;36m"),
		"white":        []byte("\x1b[1;37m"),
	}
)

//console management
func (s *Shell) startConsole() {
	for {
		s.consoleWriteStr(shellPrefixStr)
		in, err := s.readUserIn('\n')
		if err != nil {
			s.Err(err)
			continue
		}
		in = strings.Trim(in, "\n\r\t ")
		if in == "" {
			continue
		}

		cmdParts := regCmdToken.FindStringSubmatch(in)

		if len(cmdParts) == 0 {
			cmdParts = make([]string, 3)
		}

		cmdName := cmdParts[1]
		cmdArgs := cmdParts[2]

		s.consoleWrite(shellPrefixBytes)

		if cmd, ok := s.cmds[cmdName]; ok {
			s.setCMDRunning()
			cmd.run(cmdArgs)
			s.setCMDStopped()
			s.consoleWriteStr("\n")
		} else {
			s.consoleWriteStr("invalid command, use \\? to list available commands\n")
		}

	}
}

func (s *Shell) drainBuffer() {
	for {
		s.consoleMu.Lock()
		if s.consoleBuf.Len() > 0 && !s.cmdIsRunning() {
			s.c.LogFile.Write(s.consoleBuf.Bytes())
			s.consoleOut.Write(s.consoleBuf.Bytes())
			s.consoleBuf.Reset()
		}
		s.consoleMu.Unlock()
		time.Sleep(bufferDrainRate)
	}
}

func (s *Shell) readUserIn(delim byte) (string, error) {
	in, err := s.userIn.ReadString(delim)
	s.c.LogFile.Write([]byte(in))
	return in, err
}

func (s *Shell) consoleBuffer(data []byte) (int, error) {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()
	return s.consoleBuf.Write(data)
}

func (s *Shell) consoleBufferStr(data string) (int, error) {
	return s.consoleBuffer([]byte(data))
}

func (s *Shell) consoleWrite(data []byte) (int, error) {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()
	s.c.LogFile.Write(data)
	return s.consoleOut.Write(data)
}

func (s *Shell) consoleWriteStr(data string) (int, error) {
	return s.consoleWrite([]byte(data))
}

func (s *Shell) consoleWritePrefix(data []byte) (int, error) {
	return s.consoleWrite(shellPrefix(data))
}

func (s *Shell) consoleWriteStrPrefix(data string) (int, error) {
	return s.consoleWrite(shellPrefix([]byte(data)))
}

func (s *Shell) consoleBufferStrWrappedErr(sock *socket, h *header, data string) (int, error) {
	return s.consoleBufferWrappedErr(sock, h, []byte(data))
}

func (s *Shell) consoleBufferWrappedErr(sock *socket, h *header, data []byte) (int, error) {
	out, err := s.wrapSocketOutput(sock, h, data, true)
	if err != nil {
		return 0, err
	}

	return s.consoleBuffer(out)
}

func (s *Shell) consoleWriteStrWrappedErr(sock *socket, h *header, data string) (int, error) {
	return s.consoleWriteWrappedErr(sock, h, []byte(data))
}

func (s *Shell) consoleWriteWrappedErr(sock *socket, h *header, data []byte) (int, error) {
	out, err := s.wrapSocketOutput(sock, h, data, true)
	if err != nil {
		return 0, err
	}

	return s.consoleWrite(out)
}

func (s *Shell) consoleBufferStrWrapped(sock *socket, h *header, data string) (int, error) {
	return s.consoleBufferWrapped(sock, h, []byte(data))
}

func (s *Shell) consoleBufferWrapped(sock *socket, h *header, data []byte) (int, error) {
	out, err := s.wrapSocketOutput(sock, h, data, false)
	if err != nil {
		return 0, err
	}

	return s.consoleBuffer(out)
}

func (s *Shell) consoleWriteStrWrapped(sock *socket, h *header, data string) (int, error) {
	return s.consoleWriteWrapped(sock, h, []byte(data))
}

func (s *Shell) consoleWriteWrapped(sock *socket, h *header, data []byte) (int, error) {
	out, err := s.wrapSocketOutput(sock, h, data, false)
	if err != nil {
		return 0, err
	}

	return s.consoleWrite(out)
}

func (s *Shell) wrapSocketOutput(sock *socket, h *header, data []byte, errOut bool) ([]byte, error) {
	if h == nil {
		h = newEmptyHeader()
	}

	hd, err := h.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sID := sock.id()

	buf := bytes.NewBuffer(nil)
	if errOut {
		buf.Write(ansiColors["red"])
	}

	buf.WriteString("\n====== start socket: ")
	buf.WriteString(sID)
	buf.WriteString(", header: ")
	buf.Write(hd)
	buf.WriteString(" ======\n")

	if errOut {
		buf.WriteString("SOCKET ERROR: ")
	}

	buf.Write(data)
	buf.WriteString("\n======   end socket: ")
	buf.WriteString(sID)
	buf.WriteString(", header: ")
	buf.Write(hd)
	buf.WriteString(" ======\n")

	if errOut {
		buf.Write(ansiColors["none"])
	}

	return buf.Bytes(), nil
}

func shellPrefix(data []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(bytes.Replace(data, []byte{'\n'}, []byte("\n"+shellPrefixStr), -1))
	return buf.Bytes()
}

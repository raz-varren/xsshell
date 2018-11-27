package shell

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
)

type HeaderType uint8

const (
	//header types
	HTConsoleOut HeaderType = iota + 1
	HTAck
	HTErr  HeaderType = 122 //we probably don't need more than 122 header types
	HTNone HeaderType = 0

	HTLen int = 1
	IDLen     = 4
)

var (
	errNoHeaderType    = errors.New("no header type set")
	errHeaderWrongSize = errors.New("header length should be exactly: " + strconv.Itoa(HeaderLen))

	HeaderLen = base64.StdEncoding.EncodedLen(HTLen + IDLen)
)

type header struct {
	ht HeaderType
	id []byte
}

func (h *header) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(h.ht))
	buf.Write(h.id)

	l := base64.StdEncoding.EncodedLen(buf.Len())

	if l != HeaderLen {
		return nil, errHeaderWrongSize
	}

	b := make([]byte, l)
	base64.StdEncoding.Encode(b, buf.Bytes())

	return b, nil
}

func (h *header) UnmarshalBinary(data []byte) error {
	if len(data) != HeaderLen {
		return errHeaderWrongSize
	}

	d, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	h.ht = HeaderType(d[0])
	h.id = d[1:]

	return nil
}

func newHeader(ht HeaderType) *header {
	r := make([]byte, IDLen)
	io.ReadFull(rand.Reader, r)

	h := &header{
		ht: ht,
		id: r,
	}

	return h
}

func newEmptyHeader() *header {
	return &header{
		ht: HTNone,
		id: make([]byte, IDLen),
	}
}

func (h *header) String() string {
	d, err := h.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return string(d)
}

func (h *header) ID() string {
	return hex.EncodeToString(h.id)
}

func (h *header) genPayload(data []byte) ([]byte, error) {
	d, err := h.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return combineHeaderPayload(d, data), nil
}

func extractHeaderPayload(data []byte) (*header, []byte, error) {
	h := &header{}
	err := h.UnmarshalBinary(data[:HeaderLen])
	return h, data[HeaderLen:], err
}

func genPayload(ht HeaderType, data []byte) ([]byte, error) {
	h := newHeader(ht)
	d, err := h.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return combineHeaderPayload(d, data), nil
}

func combineHeaderPayload(h, data []byte) []byte {
	buf := bytes.NewBuffer(h)
	buf.Write(data)
	return buf.Bytes()
}

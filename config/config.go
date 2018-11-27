package config

import (
	"io"
)

type Config struct {
	ReadBufferSize  int
	WriteBufferSize int
	LogFile         io.Writer
	WrkDir          string
}

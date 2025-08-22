package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type RequestType uint

const (
	PutRequest RequestType = iota
)

func (r RequestType) String() string {
	switch r {
	case PutRequest:
		return "PutRequest"
	default:
		return "UnknownRequest"
	}
}

func main() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey || a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.SourceKey {
				a.Value = slog.StringValue(filepath.Base(a.Value.String()))
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	srvConf := ServiceConfig{
		entityID: 1,
		address:  "127.0.0.1:11234",
		addressTable: map[uint16]string{
			0: "127.0.0.1:11235",
		},
	}

	_ = NewEntity(1, "Server", srvConf)

	clientConf := ServiceConfig{
		entityID: 0,
		address:  "127.0.0.1:11235",
		addressTable: map[uint16]string{
			1: "127.0.0.1:11234",
		},
	}
	app := NewCFDPApp(0, "Client", clientConf)
	app.ListDirectory(1, "/tmp/foo.txt", "/remote")

	for {
		// Prevent busy loop
		time.Sleep(10000 * time.Millisecond)
	}
}

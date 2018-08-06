package util

import (
	"os"
	
	glogging "github.com/op/go-logging"
)

var format = glogging.MustStringFormatter(`%{color}%{time:2006-01-02T15:04:05.999} [%{module}] ▶ %{level:.4s} %{color:reset} %{message}`)
var backend = glogging.NewLogBackend(os.Stdout, "", 0)
var backendFormatter = glogging.NewBackendFormatter(backend, format)

func init() {
	glogging.SetBackend(backendFormatter)
}
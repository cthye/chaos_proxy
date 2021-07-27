package main

import (
	"fmt"
	"nessaj_proxy/config"
	"nessaj_proxy/server"
	"os"
	"testing"
)

func TestMainCheckRoutine(t *testing.T) {
	conf, err := config.Parse()
	err = server.RunServer(conf)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
}

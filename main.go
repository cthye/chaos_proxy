package main

import (
	"fmt"
	"nessaj_proxy/config"
	"nessaj_proxy/server"
	"nessaj_proxy/utils"
	"os"
)

func main() {
	conf, err := config.Parse()
	utils.ChkErr(err)
	
	err = server.RunServer(conf)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
}

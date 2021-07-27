package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"nessaj_proxy/config"
	"nessaj_proxy/database"
	"nessaj_proxy/utils"
	"time"
)

func SetupServer(conf *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(AuthenticationMiddleware(conf))

	r.GET("/version", versionHandler)
	r.POST("/register", AgentRegisterHandler)
	r.GET("/list", ListAllAgentsHandler)
	r.POST("/status", CheckAgentStatusHandler)
	r.GET("/pub", SendPubKeyHandler)
	r.POST("/cmd", SendCmdHandler)

	return r
}

type Res struct {
	IP           string
	VersionBytes []byte
}

func RoutineCheck() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover from RoutineCheck")
		}
	}()

	db, err := database.GetDB()
	if err != nil {
		log.Panicf("%s:%s", "database open", err)
	}
	defer db.Close()
	lst, num, err := database.ListAll(db)
	if err != nil {
		log.Panicf("%s:%s", "database list", err)
	}
	resCh := make(chan Res)
	for k, _ := range lst {
		ip := k
		go func(ip string) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recover 0")
				}
			}()
			resCh <- Res{ip, utils.ReqWithAuth("GET", fmt.Sprintf("http://%s/version", ip), nil)}
		}(ip)
	}
	for i := 0; i < num; i++ {
		var item []byte
		res := <-resCh
		if res.VersionBytes == nil {
			// log.Panic(err)
			item, err = json.Marshal(map[string]string{
				"Status":  "dead",
				"Version": "",
			})
		} else {
			item, err = json.Marshal(map[string]string{
				"Status":  "running",
				"Version": string(version),
			})
		}
		if err != nil {
			log.Panicf("%s:%s", "json marshal", err)
		}

		if err = database.Set([]byte(res.IP), item, db); err != nil {
			log.Panicf("%s:%s", "database set", err)
		}

		// *testing: wait second?
		time.Sleep(time.Second * 1)
	}
}

func Epoll() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover from Epoll")
		}
	}()
	// panic("test for epoll")
	// todo: delay
	ticker := time.NewTicker(3 * time.Second)
	fmt.Println(time.Now())
	go func() {
		// todo: test ticker panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recover 1")
			}
		}()
		// panic("test for ticker")
		for t := range ticker.C {
			fmt.Println(t)
			//panic("test for ticker")
			// need goroutine to avoid being trapped into panic 1
			go RoutineCheck()
		}
	}()
}

func RunServer(conf *config.Config) error {
	r := SetupServer(conf)
	go Epoll()
	return r.Run(fmt.Sprintf("%s:%d", conf.Host, conf.Port))
}

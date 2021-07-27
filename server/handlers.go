package server

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"nessaj_proxy/config"
	"nessaj_proxy/database"
	"nessaj_proxy/utils"
)

var (
	version string = "1.0.0"
)

func mkResp(data interface{}) map[string]interface{} {
	return gin.H{
		"Code": 0,
		"Data": data,
		"Msg":  "",
	}
}

func mkErrResp(code int, msg string) map[string]interface{} {
	return gin.H{
		"Code": code,
		"Data": "",
		"Msg":  msg,
	}
}

func versionHandler(c *gin.Context) {
	c.JSON(200, mkResp(version))
}

// Agent register
type AgentInfo struct {
	IP      string `json: "IP" binding: "required"`
	Status  string `json: "status" binding: "required"`
	Version string `json: "version" binding: "required"`
}

func AgentRegisterHandler(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}
	defer db.Close()

	var info AgentInfo
	err = c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(200, mkErrResp(1, err.Error()))
		return
	}
	infoBytes, err := json.Marshal(map[string]string{
		"Status":  info.Status,
		"Version": info.Version,
	})
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}
	if err = database.Set([]byte(info.IP), infoBytes, db); err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	ret, err := database.View([]byte(info.IP), db)
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	var retJson AgentInfo
	if err := json.Unmarshal(ret, &retJson); err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	c.JSON(200, mkResp(map[string]interface{}{
		"ip":   info.IP,
		"info": retJson,
	}))
}

func ListAllAgentsHandler(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}
	defer db.Close()
	res, total, err := database.ListAll(db)
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	resJson, err := json.Marshal(res)

	c.JSON(200, mkResp(map[string]interface{}{
		"total": total,
		"list":  string(resJson),
	}))
}

type AgentIP struct {
	IP string `json: "IP" binding: "required"`
}

func CheckAgentStatusHandler(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}
	defer db.Close()
	var ip AgentIP
	if err := c.ShouldBindJSON(&ip); err != nil {
		c.JSON(200, mkErrResp(1, err.Error()))
		return
	}

	ret, err := database.View([]byte(ip.IP), db)
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	c.JSON(200, mkResp(map[string]interface{}{
		"ip":     ip,
		"status": string(ret),
	}))
}

func SendPubKeyHandler(c *gin.Context) {
	conf, err := config.Parse()
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	key := conf.SenderKeyPair

	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	// encode pub_key into PEM format
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pub,
	}

	var writer bytes.Buffer
	if err := pem.Encode(&writer, block); err != nil {
		c.JSON(200, mkErrResp(2, err.Error()))
		return
	}

	c.JSON(200, mkResp(map[string]interface{}{
		"pub": string(writer.Bytes()),
	}))
}

type Cmd struct {
	AgentURL string                 `json:"agenturl" binding:"required"`
	Opt      string                 `json: "opt" binding: "required"`
	Op       string                 `json: "op" binding: "required"`
	Params   map[string]interface{} `json:"params"`
}

func SendCmdHandler(c *gin.Context) {
	var cmd Cmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(200, mkErrResp(1, err.Error()))
		return
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"op":     cmd.Op,
		"params": cmd.Params,
	})

	if err != nil {
		c.JSON(200, mkErrResp(1, err.Error()))
		return
	}

	respBytes := utils.ReqWithAuth("POST", fmt.Sprintf("%s/chaos/%s", "http://"+cmd.AgentURL, cmd.Opt), reqBody)
	if respBytes == nil {
		c.JSON(200, mkErrResp(1, err.Error()))
		return
	}

	c.JSON(200, mkResp(map[string]interface{}{
		"op":   cmd.Op,
		"resp": string(respBytes),
	}))
}

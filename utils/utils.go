package utils

import (
	"bytes"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"nessaj_proxy/config"
	"net"
	"net/http"
	"os"
	"time"
)

func ChkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func Find(val string, slice []string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func Auth() (string, error) {
	// ! concurrent map write
	// conf, err := config.Parse()
	// if err != nil {
	// 	return "", nil
	// }

	key := config.Conf.SenderKeyPair
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"exp": time.Now().Add(time.Second * 30),
		"iat": time.Now(),
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", nil
	}
	bearer := "Bearer " + tokenString
	return bearer, nil
}

func ReqWithAuth(method string, url string, body []byte) []byte {
	bearer, err := Auth()
	if err != nil {
		return nil
	}

	var req *http.Request
	if body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	}
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", bearer)
	if err != nil {
		return nil
	}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: netTransport,
		Timeout:   5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return respBytes
}

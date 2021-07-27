package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"nessaj/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)


func TestAuth(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error(err)
	}

	config := config.MkConfig("127.0.0.1", 1337, &key.PublicKey, false)

	ts := httptest.NewServer(SetupServer(&config))
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/version", ts.URL))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("no Authentication expect status code 400, got %d", resp.StatusCode)
	}

	// test for 400 (invalid authorization header)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/version", ts.URL), nil)
	basic := "basic"
	req.Header.Add("Authorization", basic)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error on response.\n[ERRO] - %s", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Test for valid authorization header failed, got %d with header %s", resp.StatusCode, req.Header)
	}

	// test for 400 (unexpected signing method)
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/version", ts.URL), nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})

	key_hmac := make([]byte, 256)

	tokenString, err := token.SignedString(key_hmac)
	if err != nil {
		t.Error(err)
	}

	bearer := "Bearer " + tokenString
	req.Header.Add("Authorization", bearer)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error on response.\n[ERRO] - %s", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Test for correct signing method failed, got %d with header %s", resp.StatusCode, req.Header)
	}

	// test for 403 (exp or iat missing)
	token = jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{})

	tokenString, err = token.SignedString(key)
	if err != nil {
		t.Error(err)
	}

	bearer = "Bearer " + tokenString

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/version", ts.URL), nil)

	req.Header.Add("Authorization", bearer)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error on response.\n[ERRO] - %s", err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("Test for specific claims failed, got %d with header %s", resp.StatusCode, req.Header)
	}

	// test for correct request
	token = jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, 0, 1),
		"iat": time.Now(),
	})

	tokenString, err = token.SignedString(key)
	if err != nil {
		t.Error(err)
	}

	bearer = "Bearer " + tokenString

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/version", ts.URL), nil)

	req.Header.Add("Authorization", bearer)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error on response.\n[ERRO] - %s", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Test for correct request failed, got %d with header %s", resp.StatusCode, req.Header)
	}
}

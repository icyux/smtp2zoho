package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ClientID, ClientSecret, Uid, RefreshToken, SelfMail string

	checkLock  sync.Mutex
	token      string
	expire     int64
	httpClient = &http.Client{
		Timeout: 20 * time.Second,
	}
)

func SendMail(mail *Mail) error {
	tokenPipe := make(chan string)
	go checkToken(tokenPipe)
	tkn := <-tokenPipe
	url := fmt.Sprintf(
		"https://mail.zoho.com/api/accounts/%s/messages",
		Uid,
	)
	fromAddr := SelfMail
	if mail.SenderName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", mail.SenderName, SelfMail)
	}
	body := map[string]string{
		"fromAddress": fromAddr,
		"toAddress":   mail.Recver,
		"subject":     mail.Subject,
		"content":     mail.Content,
	}
	encoded, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(encoded))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tkn))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return ErrRequestFailed
	}
	return nil
}

func checkToken(tokenPipe chan string) {
	checkLock.Lock()
	curTs := time.Now().Unix()
	if expire <= curTs {
		refreshToken()
	}
	tokenPipe <- token
	checkLock.Unlock()
}

func refreshToken() error {
	reqUrl := "https://accounts.zoho.com/oauth/v2/token"
	body := url.Values{
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {RefreshToken},
	}
	req, _ := http.NewRequest("POST", reqUrl, strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return ErrRequestFailed
	}
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		return err
	}

	var succ bool
	token, succ = data["access_token"].(string)
	if !succ {
		return ErrRequestFailed
	}
	expire = time.Now().Unix() + int64(data["expires_in"].(float64))
	return nil
}

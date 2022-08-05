package main

import (
	"log"
	"io/ioutil"
	"net"

	"gopkg.in/yaml.v3"
)

var listenAddr string

func init() {
	b, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		panic(err)
	}
	cfg := make(map[string]string)
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}

	listenAddr = cfg["smtpListen"]
	ClientID = cfg["ZohoClientID"]
	ClientSecret = cfg["ZohoClientSecret"]
	Uid = cfg["ZohoUserID"]
	RefreshToken = cfg["ZohoRefreshToken"]
	SelfMail = cfg["ZohoMailAddr"]
}

func main() {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	log.Printf("SMTP server started on \"%s\"", listenAddr)

	for {
		conn, _ := listener.Accept()
		go smtpHandler(conn)
	}
}

func smtpHandler(conn net.Conn) {
	mail, err := ParseSmtp(conn)
	if err != nil {
		log.Println(err)
		return
	}
	err = SendMail(mail)
	if err != nil {
		log.Printf("received a mail request to %s, failed to send: %s", mail.Recver, err)
	} else {
		log.Printf("received a mail request to %s, send succ", mail.Recver)
	}
}

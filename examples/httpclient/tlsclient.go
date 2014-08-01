package main

import (
        "crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
        cert, err := tls.LoadX509KeyPair("client.pem", "client.key")
        if err != nil {
                fmt.Printf("server: loadkeys: %s", err)
        }
        config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true, NextProtos: []string{"http/1.1"}}
        conn, err := tls.Dial("tcp", "127.0.0.1:4040", &config)
        if err != nil {
                fmt.Printf("client: dial: %s", err)
        }
	client := httputil.NewClientConn(conn, nil)
	req, err := http.NewRequest("GET", "http://localhost:4040/", nil)
	handle(err)
	res, err := client.Do(req)
	handle(err)
	data := make([]byte, int(res.ContentLength))
	_, err = res.Body.(io.Reader).Read(data)
	fmt.Println(string(data))
	res.Body.Close()
}

package main

import (
	"fmt"
	"net/http"
	smart_http "github.com/nileshjagnik/smart_server"
)

func main() {
	err := smart_http.ListenAndServeTLSSmart("localhost:4040", "server.pem", "server.key" , http.FileServer(http.Dir("./testdata")),smart_http.FileServer(smart_http.Dir("./testdata")))
	if err != nil {
		fmt.Println(err)
	}
}

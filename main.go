package main

import (
	"./p3"
	"log"
	"net/http"
	"os"
)

/**
usage: go build main.go
then you can chose enter the address or not
if you don't enter the address or enter the address with 6686, the server will start at 6686 as the first node
if you enter the addresses other than 6686, then the server will start as the after node
*/
func main() {
	router := p3.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}

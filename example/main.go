package main

import (
	"io"
	"net/http"
	"os"

	"httprint"
)

func hello(w http.ResponseWriter, r *http.Request) {
	httprint.Print(r, "こんにちは")
	httprint.Print(r, "さようなら")

	io.WriteString(w, "hello")
}

//noinspection GoUnhandledErrorResult
func main() {
	httprint.Enable = true
	httprint.PrintHeader = true
	httprint.Output = os.Stdout

	const addr = ":8080"

	if true {
		http.HandleFunc("/", httprint.WrapHandlerFunc(hello))
		http.ListenAndServe(addr, nil)
	} else {
		// HandlerFuncにつけてまわれないときはこっち
		http.HandleFunc("/", hello)

		server := httprint.WrapHandler(http.DefaultServeMux)
		http.ListenAndServe(addr, server)
	}
}

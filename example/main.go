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

func main() {
	httprint.Enable = true
	httprint.PrintHeader = true
	httprint.Output = os.Stdout

	http.HandleFunc("/", httprint.WrapHandler(hello))

	//noinspection ALL
	http.ListenAndServe(":8080", nil)
}

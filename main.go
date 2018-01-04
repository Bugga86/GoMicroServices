package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", upload)

}
func upload(wr http.ResponseWriter, rq *http.Request) {
	fmt.Println(rq.Method)
	if rq.Method == "POST" {
		rq.ParseMultipartForm(32 << 20)
		file, fileHdr, err := rq.FormFile("nf")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(wr, "%v", fileHdr)
		f, err := os.OpenFile("./test/"+fileHdr.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

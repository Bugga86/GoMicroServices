package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", upload)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
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
		//		fmt.Fprintf(wr, "%v", fileHdr)
		fmt.Fprintf(wr, "%v", fileHdr.Filename+"has been uploaded")
		f, err := os.OpenFile("./"+fileHdr.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	} else {

		tpl.ExecuteTemplate(wr, "index.gohtml", upload)
	}
}

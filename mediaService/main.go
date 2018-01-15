package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var tpl *template.Template

const (
	// make 2 more Files
	// 1. json file which will have this mapping
	// 2. a go program to read this file and get the values of constants
	S3_REGION = "ap-south-1"    //region name for AWS Mumbai
	S3_BUCKET = "go-case-study" //s3 bucket name
)

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", upload)
	log.Println("Server started Successfully and listening on 8080 port")
	//http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	//http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)

}
func upload(wr http.ResponseWriter, rq *http.Request) {
	//log.Println("Server started Successfully")
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}

	if rq.Method == "POST" {
		rq.ParseMultipartForm(32 << 20)
		file, fileHdr, err := rq.FormFile("nf")
		if err != nil {
			fmt.Println(err)
			return
		}
		AccountName := rq.FormValue("accountName")
		//fmt.Println(AccountName)

		//fmt.Printf("%T", file)
		defer file.Close()
		//		fmt.Fprintf(wr, "%v", fileHdr)
		fmt.Fprint(wr, "File Uploaded Successfully")
		log.Println("Copying File in to the Server")
		f, err := os.OpenFile("./imagesRcv/"+fileHdr.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		log.Println("File Successfully Copied into the Server")
		fileInfo, err := f.Stat()
		var size int64 = fileInfo.Size()
		buffer := make([]byte, size)
		file.Read(buffer)
		log.Println("Uploading file into S3 Bucket")
		log.Println("Connecting to S3...")
		_, err = s3.New(s).PutObject(&s3.PutObjectInput{
			Bucket:               aws.String(S3_BUCKET),
			Key:                  aws.String(AccountName + "/" + fileHdr.Filename),
			ACL:                  aws.String("private"),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ServerSideEncryption: aws.String("AES256"),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("File Successfully uploaded into the server")
		mapD := map[string]string{"imageName": fileHdr.Filename, "accountName": AccountName}
		s, err := json.Marshal(mapD)
		newFileName := strings.TrimSuffix(fileHdr.Filename, filepath.Ext(fileHdr.Filename))
		fmt.Println(newFileName)
		ioutil.WriteFile("./json/out/"+AccountName+newFileName+".json", s, 0644)
		log.Println("JSON file has been created and pushed into the local MQ")
	} else {

		tpl.ExecuteTemplate(wr, "index.gohtml", upload)
	}

}

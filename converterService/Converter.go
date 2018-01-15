package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nfnt/resize"
)

type Info struct {
	AccountName string
	ImageName   string
}

const (
	S3_REGION = "ap-south-1"    //region name for AWS Mumbai
	S3_BUCKET = "go-case-study" //s3 bucket name
)

func main() {
	f, err := ioutil.ReadDir("./mqDest/")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range f {
		//fmt.Println(file.Name())
		log.Println("Getting details of Image file")
		fInfo := getImageDetails(file.Name())
		log.Println("Received details by parsing json file, now connecting to S3 and downloading Image")
		log.Println("Connecting to S3...")
		n := imageDownloader(fInfo)
		log.Printf("The File has been downloaded, size %v bytes", n)
		log.Println("Creating ThumbNails of the original Image")
		err := imageThumbNail(fInfo)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Uploading thumbnails to S3 bucket in the respective Account Name")
		thumbNailImageUpload(fInfo)
	}
}
func getImageDetails(f string) Info {
	//Parsing the JSON file and getting Filename and AccountName information
	raw, err := ioutil.ReadFile("./mqDest/" + f)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//str := string(raw)
	var c Info
	err1 := json.Unmarshal(raw, &c)
	if err1 != nil {
		log.Fatal(err1)
	}
	return c
	//return c
}
func imageDownloader(fInfo Info) int64 {
	//Downloading image file from AWS S3
	Acc := fInfo.AccountName
	fName := fInfo.ImageName
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)

	f, err := os.Create("./imageProcessing/" + fName)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("dummy file created")
	//fmt.Println(Acc + "/" + fName)
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(Acc + "/" + fName),
	})

	if err != nil {
		log.Fatal(err)
	}
	return n
}

func imageThumbNail(fInfo Info) error {
	AccountName := fInfo.AccountName
	origfName := fInfo.ImageName
	imagePath, _ := os.Open("./imageProcessing/" + origfName)
	srcImage, _, _ := image.Decode(imagePath)
	newImage := resize.Thumbnail(80, 80, srcImage, resize.Lanczos3)
	out, err := os.Create("./imageProcessing/thumbnails/" + AccountName + origfName + "_T" + ".jpg")
	jpeg.Encode(out, newImage, nil)

	return err
}
func thumbNailImageUpload(fInfo Info) error {
	Acc := fInfo.AccountName
	fName := fInfo.ImageName
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile("./imageProcessing/thumbnailImages/"+fName, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	fDet, _ := f.Stat()
	var size int64 = fDet.Size()
	buffer := make([]byte, size)

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(Acc + "/" + fName),
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
	return err
}

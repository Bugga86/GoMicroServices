package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

func mqCopy(fileName string) {

	file, err := os.Open("./mediaService/json/out/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	f, err := os.OpenFile("./converterService/mqDest/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer os.Remove("./mediaService/json/out/" + fileName)
	io.Copy(f, file)
	fileInfo, _ := f.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	log.Print("File has been sent " + fileName)
}
func main() {

	isShuttingDown := false
	started := false

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

OUTER:
	for {

		if isShuttingDown == true {
			break
		}
		select {
		case _, ok := <-c:
			if ok {
				fmt.Println("Service is shutting down")
				isShuttingDown = true
				c = nil
				continue OUTER
			}

		default:
			if !started {
				fmt.Println("MQ Service Started Successfully")
				started = true
			}

			//	fmt.Print(".")
			fNames, _ := ioutil.ReadDir("./mediaService/json/out/")
			if len(fNames) == 0 {
				time.Sleep(100 * time.Millisecond)
			}
			for _, f := range fNames {
				//		fmt.Printf("%T", f.Name())
				go mqCopy(f.Name())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

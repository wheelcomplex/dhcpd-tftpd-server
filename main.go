// simple dhcp + tftp server

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pin/tftp"
)

// upload to this directory
var updir string = ""
var dirsymble string = "/"

var port int = 9080

func main() {

	var err error
	var homedir string = os.Getenv("HOME")
	var argline string = strings.Join(os.Args[1:], " ")

	if runtime.GOOS == "windows" {
		dirsymble = "\\"
	}

	flag.StringVar(&updir, "w", HOME_UPLOAD_DIR, "directory to write uploaded file")
	flag.IntVar(&port, "p", 9080, "listening port")

	flag.Parse()

	if updir == HOME_UPLOAD_DIR {
		updir = homedir + "/upload"
	}

	err = os.MkdirAll(updir, 0700)
	if err != nil {
		log.Fatalf("change to the upload directory: %s\n", err)
	}

	err = os.Chdir(updir)
	if err != nil {
		log.Fatalf("change to the upload directory: %s\n", err)
	}
	upfile, err := os.Create(updir + "/.uploader.rw.check")
	if err != nil {
		log.Printf("WARNING: write to the upload directory: %s\n", err)
	}
	upfile.Close()
	os.Remove(updir + "/.uploader.rw.check")

	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)  // optional
	err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}

	log.Printf("[%s] Listening at %s, upload to %s ...\n", strings.Trim(os.Args[0]+" "+argline, " "), strconv.Itoa(port), updir)

}

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes sent\n", n)
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes received\n", n)
	return nil
}

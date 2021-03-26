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

// root directory
var rootdir string = ""

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

	flag.StringVar(&rootdir, "w", "", "directory to write uploaded file")
	flag.StringVar(&updir, "w", "", "directory to write uploaded file")
	flag.IntVar(&port, "p", 69, "listening port")

	flag.Parse()

	if rootdir == "" {
		rootdir = homedir + dirsymble + "tftproot"
	}

	err = os.Chdir(rootdir)
	if err != nil {
		err = os.MkdirAll(rootdir, 0700)
		if err != nil {
			log.Fatalf("create the root directory: %s\n", err)
		}
		// re-try
		err = os.Chdir(rootdir)
		if err != nil {
			log.Fatalf("change to the root directory: %s\n", err)
		}
	}

	if updir == "" {
		updir = rootdir + dirsymble + "upload"
	}

	err = os.Chdir(updir)
	if err != nil {
		err = os.MkdirAll(updir, 0700)
		if err != nil {
			log.Fatalf("create the upload directory: %s\n", err)
		}
		// re-try
		err = os.Chdir(updir)
		if err != nil {
			log.Fatalf("change to the upload directory: %s\n", err)
		}
	}
	upfile, err := os.Create(updir + "/.uploader.rw.check")
	if err != nil {
		log.Printf("WARNING: write to the upload directory: %s\n", err)
	}
	upfile.Close()
	os.Remove(updir + "/.uploader.rw.check")

	log.SetPrefix(os.Args[0] + "[" + strconv.Itoa(os.Getgid()) + "] ")

	log.Printf("%s, Listening at %s, upload to %s ...\n", " "+argline, " "), strconv.Itoa(port), updir)

	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)                     // optional
	err := s.ListenAndServe(":" + strconv.Itoa(port)) // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}

	select {}

}

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	log.Printf("reading [%s] ...\n", filename)
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
	log.Printf("%d bytes sent\n", n)
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	log.Printf("writing [%s] ...\n", filename)
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
	log.Printf("%d bytes received\n", n)
	return nil
}

// simple dhcp + tftp server

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pin/tftp"
)

// root directory
var rootDir string = ""

// upload to this directory
var uploadDir string = ""
var dirsymble string = "/"

var port int = 9080

func main() {

	var err error
	var homedir string = os.Getenv("HOME")
	var argline string = strings.Join(os.Args[1:], " ")

	if runtime.GOOS == "windows" {
		dirsymble = "\\"
	}

	flag.StringVar(&rootDir, "r", "", "directory to read file")
	flag.StringVar(&uploadDir, "w", "", "directory to write uploaded file")
	flag.IntVar(&port, "p", 69, "listening port")

	flag.Parse()

	if rootDir == "" {
		rootDir = homedir + dirsymble + "tftproot"
	}

	rootDir, err = filepath.EvalSymlinks(rootDir)
	if err != nil {
		log.Fatalf("readlink the root directory: %s\n", err)
	}

	err = os.Chdir(rootDir)
	if err != nil {
		err = os.MkdirAll(rootDir, 0700)
		if err != nil {
			log.Fatalf("create the root directory: %s\n", err)
		}
		// re-try
		err = os.Chdir(rootDir)
		if err != nil {
			log.Fatalf("change to the root directory: %s\n", err)
		}
	}

	if uploadDir == "" {
		uploadDir = rootDir + dirsymble + "upload"
	}

	err = os.Chdir(uploadDir)
	if err != nil {
		err = os.MkdirAll(uploadDir, 0700)
		if err != nil {
			log.Fatalf("create the upload directory: %s\n", err)
		}
		// re-try
		err = os.Chdir(uploadDir)
		if err != nil {
			log.Fatalf("change to the upload directory: %s\n", err)
		}
	}

	uploadDir, err = filepath.EvalSymlinks(uploadDir)
	if err != nil {
		log.Fatalf("readlink the upload directory: %s\n", err)
	}

	upfile, err := os.Create(uploadDir + "/.uploader.rw.check")
	if err != nil {
		log.Printf("WARNING: write to the upload directory: %s\n", err)
	}
	upfile.Close()
	os.Remove(uploadDir + "/.uploader.rw.check")

	log.SetPrefix(os.Args[0] + "[" + strconv.Itoa(os.Getpid()) + "] ")

	log.Printf("Running [%s] ...\n", os.Args[0]+" "+argline)
	log.Printf("Listening at %s, root directory [%s], upload to [%s] ...\n", strconv.Itoa(port), rootDir, uploadDir)

	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)                    // optional
	err = s.ListenAndServe(":" + strconv.Itoa(port)) // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}

	select {}

}

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	filename = rootDir + dirsymble + filename
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
	filename = uploadDir + dirsymble + filename
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

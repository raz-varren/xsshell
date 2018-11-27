package main

import (
	"flag"
	"fmt"
	"github.com/raz-varren/xsshell/config"
	"github.com/raz-varren/xsshell/shell"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	fHost     = flag.String("host", "", "websocket listen address")
	fPort     = flag.String("port", "8234", "websocket listen port")
	fCert     = flag.String("cert", "", "ssl cert file")
	fKey      = flag.String("key", "", "ssl key file")
	fPath     = flag.String("path", "/s", "websocket connection path")
	fLogFile  = flag.String("log", "", "specify a log file to log all console communication")
	fWrkDir   = flag.String("wrkdir", "./", "working directory that will be used as the relative root path for any commands requiring user provided file paths")
	fServPath = flag.String("servpath", "/static/", "specify the base url path that you want to serve files from")
	fServDir  = flag.String("servdir", "", "specify a directory to serve files from. a file server will not be started if no directory is specified")
)

func main() {
	flag.Parse()

	//go doesn't have XOR comparison
	if (*fCert != "" && *fKey == "") || (*fKey != "" && *fCert == "") {
		log.Fatalln("both cert and key must be set to use ssl")
	}

	if *fServPath == *fPath {
		log.Fatalln("-path and -servpath cannot be the same")
	}

	logFile := ioutil.Discard
	if *fLogFile != "" {
		f, err := os.OpenFile(*fLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalln(err)
		}
		logFile = f
		defer f.Close()
	}

	c := &config.Config{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		LogFile:         logFile,
		WrkDir:          *fWrkDir,
	}

	s, err := shell.New(c)
	if err != nil {
		panic(err)
	}

	hostPort := net.JoinHostPort(*fHost, *fPort)

	http.Handle(*fPath, s)

	if *fServDir != "" {
		log.Printf("serving directory: %s, under url path: %s\n", *fServDir, *fServPath)
		http.Handle(*fServPath, http.StripPrefix(*fServPath, http.FileServer(http.Dir(*fServDir))))
	}
	fmt.Println("listening for sockets on", hostPort+",", "at url path:", *fPath)
	fmt.Println("starting console")
	fmt.Println("type \\? to list available commands")
	if *fCert != "" {
		log.Println(http.ListenAndServeTLS(hostPort, *fCert, *fKey, nil))
	} else {
		log.Println(http.ListenAndServe(hostPort, nil))
	}
}

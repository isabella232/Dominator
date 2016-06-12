package main

import (
	"flag"
	"fmt"
	"github.com/Symantec/Dominator/lib/constants"
	"github.com/Symantec/Dominator/lib/filegen"
	"github.com/Symantec/Dominator/lib/filegen/httpd"
	"github.com/Symantec/Dominator/lib/filegen/util"
	"github.com/Symantec/Dominator/lib/fsutil"
	"github.com/Symantec/Dominator/lib/logbuf"
	"github.com/Symantec/Dominator/lib/srpc/setupserver"
	"github.com/Symantec/tricorder/go/tricorder"
	"log"
	"os"
	"syscall"
)

var (
	configFile = flag.String("configFile", "/var/lib/filegen-server/config",
		"Name of file containing the configuration")
	logbufLines = flag.Uint("logbufLines", 1024,
		"Number of lines to store in the log buffer")
	permitInsecureMode = flag.Bool("permitInsecureMode", false,
		"If true, run in insecure mode. This gives remote access to all")
	portNum = flag.Uint("portNum", constants.BasicFileGenServerPortNumber,
		"Port number to allocate and listen on for HTTP/RPC")
)

func printUsage() {
	fmt.Fprintln(os.Stderr,
		"Usage: filegen-server [flags...] directory...")
	fmt.Fprintln(os.Stderr, "Common flags:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "directory: tree of source files")
}

func main() {
	flag.Usage = printUsage
	flag.Parse()
	tricorder.RegisterFlags()
	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "Do not run the filegen server as root")
		os.Exit(1)
	}
	circularBuffer := logbuf.New(*logbufLines)
	logger := log.New(circularBuffer, "", log.LstdFlags)
	if err := setupserver.SetupTls(); err != nil {
		logger.Println(err)
		circularBuffer.Flush()
		if !*permitInsecureMode {
			os.Exit(1)
		}
	}
	manager := filegen.New(logger)
	if *configFile != "" {
		if err := util.LoadConfiguration(manager, *configFile); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		ch := fsutil.WatchFile(*configFile, nil)
		(<-ch).Close() // Drain the first event.
		go func() {
			<-ch
			err := syscall.Exec(os.Args[0], os.Args, os.Environ())
			if err != nil {
				logger.Printf("Unable to Exec:%s\t%s\n", os.Args[0], err)
			}
		}()
	}
	httpd.AddHtmlWriter(manager)
	httpd.AddHtmlWriter(circularBuffer)
	for _, pathname := range flag.Args() {
		if err := registerSourceDirectory(manager, pathname, "/"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := httpd.StartServer(*portNum, manager, false); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create http server\t%s\n", err)
		os.Exit(1)
	}
}

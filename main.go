// wscat
//
// Usage:
//
//      wscat --connect ws://echo.websocket.org
//      wscat --connect wss://echo.websocket.org
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bcho/wscat/client"
	"github.com/chzyer/readline"
)

var (
	connect string
)

type Writer interface {
	Write([]byte) error
}

func main() {
	flag.Parse()

	rl, err := readline.New("< ")
	if err != nil {
		log.Fatal(err)
	}

	var writer Writer

	if connect != "" {
		c, err := client.NewClient(connect, client.WithStdout(rl))
		if err != nil {
			log.Fatal(err)
		}
		go c.ConnectAndServe()
		writer = c
	} else {
		showUsage()
		os.Exit(1)
	}

	for {
		line, err := rl.Readline()
		if err != nil {
			// TODO handle error
			os.Exit(0)
		}

		go func() {
			if err := writer.Write([]byte(line)); err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func showUsage() {
	fmt.Println("usage: wscat --connect <url>")
}

func init() {
	flag.StringVar(&connect, "connect", "", "connect to a websocket server")
}

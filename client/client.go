// wscat websocket client
package client

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var (
	ErrClosed = errors.New("closed")
)

type Client interface {
	Write([]byte) error
	ConnectAndServe() error
	Close()
}

type client struct {
	url *url.URL

	closed    bool
	chanWrite chan []byte
	chanDone  chan struct{}

	dialer *websocket.Dialer
	stdout io.Writer
}

type clientOptSetter func(*client)

func WithWebsocketDialer(dialer *websocket.Dialer) clientOptSetter {
	return func(c *client) {
		c.dialer = dialer
	}
}

func WithStdout(output io.Writer) clientOptSetter {
	return func(c *client) {
		c.stdout = output
	}
}

func NewClient(rawurl string, optionSetters ...clientOptSetter) (Client, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	c := &client{
		url: u,

		closed:    false,
		chanWrite: make(chan []byte),
		chanDone:  make(chan struct{}),

		dialer: websocket.DefaultDialer,
		stdout: os.Stdout,
	}

	for _, setter := range optionSetters {
		setter(c)
	}

	return c, nil
}

func (c *client) ConnectAndServe() error {
	conn, _, err := c.dialer.Dial(c.url.String(), nil)
	if err != nil {
		return err
	}
	chanErr := make(chan error)

	defer func() {
		defer conn.Close()
		defer close(chanErr)
		c.closed = true
	}()

	go func() {
		for {
			select {
			case <-c.chanDone:
				return
			case message := <-c.chanWrite:
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					chanErr <- err
					return
				}
			}
		}
	}()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				chanErr <- err
				return
			}
			fmt.Fprintf(c.stdout, "> %s\n", message)
		}
	}()

	select {
	case <-c.chanDone:
		return nil
	case err := <-chanErr:
		return err
	}
}

func (c client) Write(message []byte) error {
	if c.closed {
		return ErrClosed
	}

	c.chanWrite <- message
	return nil
}

func (c *client) Close() {
	close(c.chanDone)
}

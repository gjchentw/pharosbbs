package bbs

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PatrickRudolph/telnet"
	"github.com/PatrickRudolph/telnet/options"
	"github.com/gin-gonic/gin"

	"github.com/ginmills/ginmill"
	"github.com/pharosrocks/pharosbbs/websocket"
)

func (s *Server) bbsd() *ginmill.Features {
	r := gin.New()
	r.GET("/", gin.HandlerFunc(s.telnetHandler))

	return ginmill.NewFeatures(r.Routes())
}

func (s *Server) telnetHandler(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)

	defer conn.Close()

	options := []telnet.Option{
		options.TerminalTypeOption,
		options.NAWSOption,
		options.EchoOption,
		options.SuppressGoAheadOption,
		options.BinaryTransmissionOption,
	}

	telnetConn := telnet.NewConnection(conn, options)
	handler := telnet.HandleFunc(exampleHandler)
	handler.HandleTelnet(telnetConn)

}

func TelnetRoutine(c *telnet.Connection) {
	exampleHandler(c)

}

func exampleHandler(c *telnet.Connection) {
	log.Printf("Connection received: %s", c.RemoteAddr())
	lr := NewReader()
	go lr.Read(c)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for line := range lr.C {
			log.Printf("Received line: %v", string(line))
		}
	}()
	time.Sleep(time.Millisecond)
	nh := c.OptionHandlers[telnet.TeloptNAWS].(*options.NAWSHandler)
	log.Printf("Client width: %d, height: %d", nh.Width, nh.Height)
	wg.Wait()
	log.Printf("Goodbye %s!", c.RemoteAddr())
}

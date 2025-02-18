package main

import (
	"andrewka/chatclient/message"
	"andrewka/chatclient/tui"
	"crypto/tls"
	"errors"
	"flag"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"net"
)

var p *tea.Program
var address = flag.String("address", "", "Адрес чат-сервера")

func main() {
	flag.Parse()
	if *address == "" {
		flag.PrintDefaults()
		return
	}

	conn, err := newConn(*address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p = tui.NewProgram(conn)

	go func() {
		handleConn(conn)
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal("Произошла ошибка во время выполнения программы:", err)
	}
}

func newConn(address string) (net.Conn, error) {
	config := tls.Config{}
	return tls.Dial("tcp", address, &config)
}

// handleConn читает поток и отправляет входящие сообщения tui-приложению.
func handleConn(conn net.Conn) {
	d := message.Decoder{conn}
	msgCh := make(chan message.Msg)

	go func() {
		for msg := range msgCh {
			p.Send(tui.OuterMsg(msg))
		}
	}()

	if err := d.Decode(msgCh); err != nil {
		p.Send(tui.ErrMsg{
			Err:   errors.New("Ошибка при подключении к серверу. Выход из программы..."),
			Fatal: true,
		})
	}
}

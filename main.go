package main

import (
	"net"
	"log"
	tea "github.com/charmbracelet/bubbletea"
	"errors"
	"andrewka/chatclient/tui"
	"andrewka/chatclient/message"
)

var p *tea.Program

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p = tui.NewProgram(conn)

	go func() {
		handleConn(conn)
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal("error running program:", err)
	}
}

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
			Err: errors.New("Ошибка при подключении к серверу. Выход из программы..."),
			Fatal: true,
		})
	}
}
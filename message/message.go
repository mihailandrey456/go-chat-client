package message

type Msg struct {
	From string
	Content string
}

func (m Msg) String() string {
	return m.From + ": " + m.Content
}
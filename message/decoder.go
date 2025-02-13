package message

import(
	"io"
	"encoding/json"
	"errors"
)

type Decoder struct {
	Reader io.Reader
}

func (dec Decoder) Decode(msgCh chan<- Msg) error {
	defer func() {
		close(msgCh)
	}()

	d := json.NewDecoder(dec.Reader)
	var msg Msg
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		if t == json.Delim('{') {
			msg = Msg{}
			continue
		} else if t == json.Delim('}') {
			msgCh <- msg
			continue
		}
		
		t, ok := t.(string)
		if !ok {
			return errors.New("unexpected key type:")
		}
		if t != "From" && t != "Content" {
			if err := skip(d); err != nil {
				return err
			}
			continue
		}

		if t == "From" {
			t, err := d.Token()
			if err != nil {
				return err
			}
			v, ok := t.(string)
			if !ok {
				return errors.New("unexpected value type")
			}
			msg.From = v
		} else if t == "Content" {
			t, err := d.Token()
			if err != nil {
				return err
			}
			v, ok := t.(string)
			if !ok {
				return errors.New("unexpected value type")
			}
			msg.Content = v
		}
	}
	return nil
}

// skip skips the next value in the JSON document.
func skip(d *json.Decoder) error {
	n := 0
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch t {
		case json.Delim('['), json.Delim('{'):
			n++
		case json.Delim(']'), json.Delim('}'):
			n--
		}
		if n == 0 {
			return nil
		}
	}
}
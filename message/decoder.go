package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Декодирует Msg из потока.
type Decoder struct {
	Reader io.Reader
}

func (dec Decoder) Decode(msgCh chan<- Msg) error {
	defer func() {
		close(msgCh)
	}()

	d := json.NewDecoder(dec.Reader)
	for {
		if err := expect(d, json.Delim('{')); err != nil {
			return err
		}

		if err := expect(d, "From"); err != nil {
			return err
		}

		from, err := extractString(d)
		if err != nil {
			return err
		}

		if err := expect(d, "Content"); err != nil {
			return err
		}

		content, err := extractString(d)
		if err != nil {
			return err
		}

		if err := expect(d, json.Delim('}')); err != nil {
			return err
		}

		msgCh <- Msg{from, content}
	}
	return nil
}

// expect возращает ошибку, если следующий токен не является expected.
func expect(d *json.Decoder, expected interface{}) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t != expected {
		return fmt.Errorf("Получен %v, ожидался %v", t, expected)
	}
	return nil
}

func extractString(d *json.Decoder) (string, error) {
	t, err := d.Token()
	if err != nil {
		return "", err
	}
	v, ok := t.(string)
	if !ok {
		return "", errors.New("Неожиданный тип токена")
	}
	return v, nil
}

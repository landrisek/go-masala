package masala

import (
	"bytes"
	"net/url"
)

type IMessage interface {
	Data(parameters url.Values, buffer *bytes.Buffer)
	Id() string
}
//go:build !yq_nojson

package yqlib

import (
	"bytes"
	"encoding/json"
	"io"

	goccyjson "github.com/goccy/go-json"
)

type jsonDecoder struct {
	decoder *json.Decoder
}

func NewJSONDecoder() Decoder {
	return &jsonDecoder{}
}

func (dec *jsonDecoder) Init(reader io.Reader) error {
	dec.decoder = json.NewDecoder(reader)
	return nil
}

func (dec *jsonDecoder) Decode() (*CandidateNode, error) {
	// encoding/json's RawMessage decode runs the full syntax scanner, so it
	// rejects things like a missing comma between object members/array
	// elements that goccy's streaming token API otherwise lets through.
	var raw json.RawMessage
	if err := dec.decoder.Decode(&raw); err != nil {
		return nil, err
	}

	var dataBucket CandidateNode
	if err := goccyjson.NewDecoder(bytes.NewReader(raw)).Decode(&dataBucket); err != nil {
		return nil, err
	}

	return &dataBucket, nil
}

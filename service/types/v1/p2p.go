package v1

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// Wrap implements the p2p Wrapper interface and wraps a state sync proto message.
func (m *Message) Wrap(pb proto.Message) error {
	switch msg := pb.(type) {
	case *SecondSpRequest:
		m.Inner = &Message_SspRequest{SspRequest: msg}
	case *SecondSpRefuse:
		m.Inner = &Message_SspRefuse{SspRefuse: msg}
	case *SecondSpAck:
		m.Inner = &Message_SspAck{SspAck: msg}
	case *SecondSpManifest:
		m.Inner = &Message_SspManifest{SspManifest: msg}
	default:
		return fmt.Errorf("unknown message: %T", msg)
	}

	return nil
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped state sync
// proto message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Inner.(type) {
	case *Message_SspRequest:
		return m.GetSspRequest(), nil
	case *Message_SspRefuse:
		return m.GetSspRefuse(), nil
	case *Message_SspAck:
		return m.GetSspAck(), nil
	case *Message_SspManifest:
		return m.GetSspManifest(), nil
	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}

// Validate validates the message returning an error upon failure.
func (m *Message) Validate() error {
	if m == nil {
		return errors.New("message cannot be nil")
	}

	//TODO: validate messages
	switch msg := m.Inner.(type) {
	case *Message_SspRequest:
		return nil
	case *Message_SspRefuse:
		return nil
	case *Message_SspAck:
		return nil
	case *Message_SspManifest:
		return nil
	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}

	return nil
}

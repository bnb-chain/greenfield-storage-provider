package v1

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// Wrap implements the p2p Wrapper interface and wraps a state sync proto message.
func (m *SPMessage) Wrap(pb proto.Message) error {
	switch msg := pb.(type) {
	case *AskApprovalRequest:
		m.Inner = &SPMessage_AskApprovalRequest{AskApprovalRequest: msg}
	case *AckApproval:
		m.Inner = &SPMessage_AckApproval{AckApproval: msg}
	case *RefuseApproval:
		m.Inner = &SPMessage_RefuseApproval{RefuseApproval: msg}
	default:
		return fmt.Errorf("unknown message: %T", msg)
	}

	return nil
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped state sync
// proto message.
func (m *SPMessage) Unwrap() (proto.Message, error) {
	switch msg := m.Inner.(type) {
	case *SPMessage_AskApprovalRequest:
		return m.GetAskApprovalRequest(), nil
	case *SPMessage_AckApproval:
		return m.GetAckApproval(), nil
	case *SPMessage_RefuseApproval:
		return m.GetRefuseApproval(), nil
	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}

// Validate validates the message returning an error upon failure.
func (m *SPMessage) Validate() error {
	if m == nil {
		return errors.New("message cannot be nil")
	}

	//TODO: validate messages
	switch msg := m.Inner.(type) {
	case *SPMessage_AskApprovalRequest:
		return nil
	case *SPMessage_AckApproval:
		return nil
	case *SPMessage_RefuseApproval:
		return nil
	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}

	return nil
}

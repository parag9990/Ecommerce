package grpcweb

import "fmt"

type rawMessage struct {
	data []byte
}

type rawCodec struct{}

func (rawCodec) Name() string {
	return "proto"
}

func (rawCodec) Marshal(value any) ([]byte, error) {
	switch message := value.(type) {
	case *rawMessage:
		return message.data, nil
	case rawMessage:
		return message.data, nil
	default:
		return nil, fmt.Errorf("grpc-web proxy codec cannot marshal %T", value)
	}
}

func (rawCodec) Unmarshal(data []byte, value any) error {
	message, ok := value.(*rawMessage)
	if !ok {
		return fmt.Errorf("grpc-web proxy codec cannot unmarshal into %T", value)
	}
	message.data = append(message.data[:0], data...)
	return nil
}

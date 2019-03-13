package libproto

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

func UnmarshalBytes(messageType string, messageBytes []byte) (proto.Message, error) {
	protoType := proto.MessageType(messageType)
	if protoType == nil {
		return nil, fmt.Errorf("unknown message type: %v", messageType)
	}
	message := reflect.New(protoType.Elem()).Interface().(proto.Message)
	err := proto.Unmarshal(messageBytes, message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", messageType, err)
	}
	return message, nil
}

package main

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtobufMessage представляет любое protobuf сообщение
type ProtobufMessage interface {
	proto.Message
}

// encodeProtobufMessage кодирует protobuf сообщение в бинарный формат
func encodeProtobufMessage(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// decodeProtobufMessage декодирует бинарные данные в protobuf сообщение
func decodeProtobufMessage(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

// encodeProtobufJSON кодирует protobuf сообщение в JSON (для отладки)
func encodeProtobufJSON(msg proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	data, err := marshaler.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

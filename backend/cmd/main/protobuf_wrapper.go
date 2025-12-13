package main

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/encoding/protojson"
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
	return marshaler.Marshal(msg)
}


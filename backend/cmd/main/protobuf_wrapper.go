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
//
//lint:ignore U1000 Используется для отладки и тестирования
func encodeProtobufMessage(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// decodeProtobufMessage декодирует бинарные данные в protobuf сообщение
//
//lint:ignore U1000 Используется для отладки и тестирования
func decodeProtobufMessage(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

// encodeProtobufJSON кодирует protobuf сообщение в JSON (для отладки)
//
//lint:ignore U1000 Используется для отладки и тестирования
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

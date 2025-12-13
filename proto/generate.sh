#!/bin/bash

# Генерация Go кода
echo "Генерация Go кода из .proto файлов..."
mkdir -p backend/proto
protoc --go_out=backend/proto --go_opt=paths=source_relative \
  --proto_path=proto \
  proto/messages.proto

# Генерация TypeScript кода с помощью protobufjs
echo "Генерация TypeScript кода из .proto файлов..."
cd frontend
mkdir -p src/proto
npx pbts -o src/proto/messages.d.ts ../proto/messages.proto
npx pbjs -t static-module -w es6 -o src/proto/messages.js ../proto/messages.proto
cd ..

echo "Генерация завершена!"


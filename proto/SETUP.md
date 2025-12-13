# Настройка Protocol Buffers

## Установка protoc

### Windows
1. Скачайте protoc с https://github.com/protocolbuffers/protobuf/releases
2. Распакуйте и добавьте `bin` директорию в PATH
3. Или используйте Chocolatey: `choco install protoc`

### Linux
```bash
sudo apt-get install protobuf-compiler
```

### macOS
```bash
brew install protobuf
```

## Установка генераторов кода

### Go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### TypeScript (protobufjs уже установлен)
```bash
cd frontend
npm install --save protobufjs
```

## Генерация кода

После установки protoc выполните:

```bash
# Генерация Go кода
protoc --go_out=backend/proto --go_opt=paths=source_relative \
  --proto_path=proto \
  proto/messages.proto

# Копирование .proto файла в public для фронтенда
cp proto/messages.proto frontend/public/messages.proto

# Или используйте скрипт (после установки protoc):
chmod +x proto/generate.sh
./proto/generate.sh
```

## После генерации

Код будет автоматически использовать protobuf вместо бинарного формата.


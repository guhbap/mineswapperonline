# Protocol Buffers для WebSocket сообщений

## Установка зависимостей

### Go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### TypeScript/JavaScript
```bash
cd frontend
npm install --save protobufjs
npm install --save-dev @types/protobufjs
```

## Генерация кода

Запустите скрипт генерации:
```bash
chmod +x proto/generate.sh
./proto/generate.sh
```

Или вручную:

### Go
```bash
protoc --go_out=backend/proto --go_opt=paths=source_relative \
  proto/messages.proto
```

### TypeScript
```bash
cd frontend
npx pbjs -t static-module -w es6 -o src/proto/messages.js ../proto/messages.proto
npx pbts -o src/proto/messages.d.ts src/proto/messages.js
```


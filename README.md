# File gRPC Service

gRPC сервис для загрузки, скачивания и просмотра файлов с ограничением конкурентных запросов.

## Возможности

- Загрузка бинарных файлов (изображений) стримингом (client‑streaming)
- Скачивание файлов стримингом (server‑streaming)
- Просмотр списка всех загруженных файлов с метаданными:
  - Имя файла
  - Дата создания
  - Дата обновления
  - Размер
- Ограничение одновременных подключений:
  - Upload/Download – **10** конкурентных запросов
  - ListFiles – **100** конкурентных запросов
- Чистая архитектура с разделением на слои
- Unit‑тесты с высоким покрытием

## Технологии

- Go 1.25+
- gRPC
- Protocol Buffers
- Testify (тестирование)
- Docker

## Быстрый старт

### Требования

- Установленный Go (версия 1.25)
- Установленный protoc и плагины (см. `make deps`)

### Установка и запуск через Docker

```bash
docker run -d -p 50051:50051 --name file_grpc golang:1.25 bash -c "apt update && apt install -y make protobuf-compiler && rm -rf file_grpc && git clone https://github.com/Hiddan13/file_grpc.git && cd file_grpc && make deps proto build && ./bin/server"

### Дождаться завершения установки

docker exec -it file_grpc bash 

chmod +x file_grpc/testing.sh

./file_grpc/testing.sh

### Установка и запуск bash 

```bash
# Клонируйте репозиторий
git clone https://github.com/Hiddan13/github.com/Hiddan13/file_grpc.git
cd file-grpc

# Установите зависимости и сгенерируйте protobuf
make deps
make proto

# Соберите сервер и клиент
make build

# Запустите сервер
./bin/server
# или
make run-server

# В другом терминале – примеры клиента
./bin/client -action upload -file test.jpg
./bin/client -action list
./bin/client -action download -file test.jpg
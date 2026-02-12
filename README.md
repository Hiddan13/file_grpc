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

## Быстрый старт

### Требования

- Установленный Go (версия 1.25)
- Установленный protoc и плагины (см. `make deps`)

### Установка и запуск

```bash
# Клонируйте репозиторий
git clone https://github.com/Hiddan13/file_grpc.git
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
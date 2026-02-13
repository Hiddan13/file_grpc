# File gRPC Service

gRPC сервис для загрузки, скачивания и просмотра файлов с ограничением конкурентных запросов.

## Возможности

- Загрузка бинарных файлов стримингом 
- Скачивание файлов стримингом 
- Просмотр списка всех загруженных файлов с метаданными:
  - Имя файла
  - Дата создания
  - Дата обновления
  - Размер
- Ограничение одновременных подключений:
  - Upload/Download – **10** конкурентных запросов
  - ListFiles – **100** конкурентных запросов
- Unit‑тесты 

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

chmod +x file_grpc/testing.sh ### Будем запускать тест создания и сохранения файла + скачивание файла 

cd file_grpc
ls my_test_repo ### ничего пока нет
./testing.sh 
ls my_test_repo ### есть файл из тестов
ls

### Установка и запуск bash 

```bash
# Запустим докер 
docker run -it --rm --name file_grpc_test golang:1.25 bash
# Установим необходимые пакеты
apt update && apt install -y make protobuf-compiler git
# Клонируйте репозиторий
git clone https://github.com/Hiddan13/file_grpc.git
cd file_grpc

# Создаём тестовый файл 
echo "Тест загрузки gRPC из тестового Docker" > test.txt

# Установите зависимости и сгенерируйте protobuf соберем сервер и клиент
make deps proto build


# Запустите сервер в фоне
./bin/server &

# Загружаем файл
./bin/client -action upload -file test.txt

# Получаем список файлов
./bin/client -action list

# Скачиваем файл
./bin/client -action download -file test.txt

# Проверяем содержимое
cat downloaded_test.txt

# Убеждаемся, что файл сохранён на сервере
ls -la my_test_repo/


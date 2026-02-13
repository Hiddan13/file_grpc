#!/bin/bash
set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "==================================="
echo "🚀 Файловый gRPC сервис – интеграционный тест"
echo "==================================="

# 1. Создаём тестовый файл
echo "[1/7] Создание тестового файла..."
echo "Hello gRPC from test script" > test.jpg
TEST_CONTENT="Hello gRPC from test script"
echo "✅ Тестовый файл создан: test.jpg (${#TEST_CONTENT} байт)"

# # 2. Запускаем сервер в фоне
# echo "[2/7] Запуск gRPC сервера..."
# ./bin/server > server.log 2>&1 &
# SERVER_PID=$!
# sleep 2  # даём серверу время подняться
# if kill -0 $SERVER_PID 2>/dev/null; then
#     echo "✅ Сервер запущен (PID: $SERVER_PID)"
# else
#     echo -e "${RED}❌ Сервер не запустился${NC}"
#     cat server.log
#     exit 1
# fi

# 3. Загружаем файл
echo "[3/7] Загрузка файла..."
UPLOAD_OUTPUT=$(./bin/client -action upload -file test.jpg 2>&1)
if echo "$UPLOAD_OUTPUT" | grep -q "Uploaded:"; then
    echo "✅ Файл успешно загружен"
else
    echo -e "${RED}❌ Ошибка загрузки${NC}"
    echo "$UPLOAD_OUTPUT"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# 4. Получаем список файлов
echo "[4/7] Запрос списка файлов..."
LIST_OUTPUT=$(./bin/client -action list 2>&1)
if echo "$LIST_OUTPUT" | grep -q "test.jpg"; then
    echo "✅ Файл присутствует в списке"
else
    echo -e "${RED}❌ Файл не найден в списке${NC}"
    echo "$LIST_OUTPUT"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# 5. Скачиваем файл
echo "[5/7] Скачивание файла..."
DOWNLOAD_OUTPUT=$(./bin/client -action download -file test.jpg 2>&1)
if echo "$DOWNLOAD_OUTPUT" | grep -q "Downloaded"; then
    echo "✅ Файл успешно скачан"
else
    echo -e "${RED}❌ Ошибка скачивания${NC}"
    echo "$DOWNLOAD_OUTPUT"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# 6. Проверяем содержимое скачанного файла
echo "[6/7] Проверка содержимого..."
DOWNLOADED_CONTENT=$(cat downloaded_test.jpg)
if [ "$DOWNLOADED_CONTENT" = "$TEST_CONTENT" ]; then
    echo "✅ Содержимое совпадает"
else
    echo -e "${RED}❌ Содержимое не совпадает${NC}"
    echo "Ожидалось: $TEST_CONTENT"
    echo "Получено: $DOWNLOADED_CONTENT"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# 7. Проверяем, что файл сохранился на диске сервера
echo "[7/7] Проверка сохранения на диск..."
if [ -f "uploads/test.jpg" ]; then
    SERVER_SIZE=$(stat -c%s "uploads/test.jpg" 2>/dev/null || stat -f%z "uploads/test.jpg" 2>/dev/null)
    echo "✅ Файл сохранён на сервере (размер: $SERVER_SIZE байт)"
else
    echo -e "${RED}❌ Файл не найден в uploads/ ${NC}"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Убиваем сервер
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null || true

echo "==================================="
echo -e "${GREEN}✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО${NC}"
echo "==================================="
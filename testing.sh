#!/bin/bash
set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "==================================="
echo "🚀 Файловый gRPC сервис – интеграционный тест (предполагается запущенный сервер)"
echo "==================================="

# 1. Создаём тестовый файл
echo "[1/6] Создание тестового файла..."
echo "Hello gRPC from test script" > test.jpg
TEST_CONTENT="Hello gRPC from test script"
echo "✅ Тестовый файл создан: test.jpg (${#TEST_CONTENT} байт)"

# 2. Загружаем файл
echo "[2/6] Загрузка файла..."
UPLOAD_OUTPUT=$(./bin/client -action upload -file test.jpg 2>&1)
if echo "$UPLOAD_OUTPUT" | grep -q "Uploaded:"; then
    echo "✅ Файл успешно загружен"
else
    echo -e "${RED}❌ Ошибка загрузки${NC}"
    echo "$UPLOAD_OUTPUT"
    exit 1
fi

# 3. Получаем список файлов
echo "[3/6] Запрос списка файлов..."
LIST_OUTPUT=$(./bin/client -action list 2>&1)
if echo "$LIST_OUTPUT" | grep -q "test.jpg"; then
    echo "✅ Файл присутствует в списке"
else
    echo -e "${RED}❌ Файл не найден в списке${NC}"
    echo "$LIST_OUTPUT"
    exit 1
fi

# 4. Скачиваем файл
echo "[4/6] Скачивание файла..."
DOWNLOAD_OUTPUT=$(./bin/client -action download -file test.jpg 2>&1)
if echo "$DOWNLOAD_OUTPUT" | grep -q "Downloaded"; then
    echo "✅ Файл успешно скачан"
else
    echo -e "${RED}❌ Ошибка скачивания${NC}"
    echo "$DOWNLOAD_OUTPUT"
    exit 1
fi

# 5. Проверяем содержимое скачанного файла
echo "[5/6] Проверка содержимого..."
DOWNLOADED_CONTENT=$(cat downloaded_test.jpg)
if [ "$DOWNLOADED_CONTENT" = "$TEST_CONTENT" ]; then
    echo "✅ Содержимое совпадает"
else
    echo -e "${RED}❌ Содержимое не совпадает${NC}"
    echo "Ожидалось: $TEST_CONTENT"
    echo "Получено: $DOWNLOADED_CONTENT"
    exit 1
fi

# 6. Проверяем, что файл сохранился на диске сервера
echo "[6/6] Проверка сохранения на диск..."
if [ -f "my_test_repo/test.jpg" ]; then
    SERVER_SIZE=$(stat -c%s "my_test_repo/test.jpg" 2>/dev/null || stat -f%z "my_test_repo/test.jpg" 2>/dev/null)
    echo "✅ Файл сохранён на сервере (размер: $SERVER_SIZE байт)"
else
    echo -e "${RED}❌ Файл не найден в uploads/ ${NC}"
    exit 1
fi

echo "==================================="
echo -e "${GREEN}✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО${NC}"
echo "==================================="
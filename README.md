# XrayDumper

Небольшая CLI-утилита на Go для получения ссылок из подписок.



## Требования

Скачайте и установите [go](https://go.dev/dl/)

## Сборка

Для сборки под Linux:

```bash
chmod +x build.sh
./build.sh
```

или

```bash
chmod +x build.sh
sh build.sh
```

Готовый бинарник появится в:

```text
output/xd-linux
```

Для Windows запустите файл:

```bat
build.bat
```

Готовый `.exe` будет находиться в:

```text
output/xd.exe
```

Также можно собрать вручную:

```bash
go build -o output/xd
```

Для Windows:

```bash
go build -o output/xd.exe
```

## Использование

Укажите ссылку на подписку:

```bash
./xd-linux -url "https://example.com/subscription"
```

Или запустите программу без `-url` — она попросит ввести ссылку в консоли.

По умолчанию найденные ссылки сохраняются в `links.txt`.

Можно указать свой файл:

```bash
./xd-linux -url "https://example.com/subscription" -out configs.txt
```

Также поддерживается `HWID`:

```bash
./xd-linux -url "https://example.com/subscription" -hwid "your-hwid"
```

Просмотр справки:

```bash
./xd-linux -help
```


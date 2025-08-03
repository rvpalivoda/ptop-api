# PTOP API

Rest api на golang, gin, postgresql, GORN ORM.

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `DB_DSN` | строка подключения к базе данных |
| `PORT` | порт HTTP-сервера (по умолчанию 8080) |
| `BTC_RPC_HOST` | адрес Bitcoin RPC |
| `BTC_RPC_USER` | логин Bitcoin RPC |
| `BTC_RPC_PASS` | пароль Bitcoin RPC |
| `ETH_RPC_URL` | URL Ethereum RPC |
| `MONERO_RPC_URL` | URL Monero RPC |
| `REDIS_ADDR` | адрес сервера Redis |
| `REDIS_PASSWORD` | пароль Redis (если требуется) |
| `REDIS_DB` | номер базы Redis |
| `CHAT_CACHE_LIMIT` | количество сообщений истории в кешe |
| `S3_ENDPOINT` | адрес S3/MinIO |
| `S3_ACCESS_KEY` | ключ доступа S3/MinIO |
| `S3_SECRET_KEY` | секретный ключ S3/MinIO |
| `S3_BUCKET` | имя бакета |
| `S3_REGION` | регион S3 |
| `S3_USE_SSL` | использовать HTTPS при подключении |

## WebSocket чат ордера

Подписка на обновления сообщений осуществляется через WebSocket:

```
wss://<host>/ws/orders/{orderID}/chat
```

Перед подключением клиент должен получить `access_token` и передать его в заголовке `Authorization: Bearer <token>`.

После подключения сервер отправит историю последних сообщений из кеша Redis. Чтобы отправить новое сообщение, нужно послать JSON:

```json
{ "content": "текст сообщения" }
```

Каждое отправленное сообщение будет сохранено в БД и рассылается всем подключённым участникам ордера.

## MinIO для хранения файлов

### Запуск

Простейший способ поднять MinIO локально — через `docker-compose`:

```yaml
version: '3.7'
services:
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
```

Запускаем:

```bash
docker-compose up -d
```

### Настройка переменных окружения

В `.env` указываем параметры доступа:

```env
S3_ENDPOINT=127.0.0.1:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=ptop
S3_REGION=us-east-1
S3_USE_SSL=false
```

### Пример загрузки и получения URL

```go
svc, _ := storage.New(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
f, _ := os.Open("example.png")
defer f.Close()
info, _ := f.Stat()
svc.Upload(context.Background(), "example.png", f, info.Size(), "image/png")
url, _ := svc.GetURL(context.Background(), "example.png", time.Hour)
fmt.Println(url)
```

Файл будет загружен в MinIO, а `url` будет содержать временную ссылку для скачивания.


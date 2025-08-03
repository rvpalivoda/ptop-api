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


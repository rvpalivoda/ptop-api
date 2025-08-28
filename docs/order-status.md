# OrderStatus: процесс и реализация

Документ описывает целевую бизнес-логику состояний ордера, разрешённые переходы, роли и права, побочные эффекты (эскроу, уведомления, WebSocket), а также план API‑эндпоинтов, требования к тестам и Swagger.

## Статусы ордера

- WAIT_PAYMENT: ордер создан, ожидается оплата покупателем/подтверждение условий сделки.
- PAID: покупатель заявил об оплате (или система подтвердила факт поступления платежа).
- RELEASED: продавец выпустил средства из эскроу покупателю — успешное завершение сделки.
- CANCELLED: ордер отменён (истёк, отменён до оплаты, либо решением арбитража в пользу продавца).
- DISPUTE: открыт спор — исход определяет арбитраж.

## Переходы (FSM)

- CREATE → WAIT_PAYMENT (создание ордера)
- WAIT_PAYMENT → PAID (подтверждение оплаты покупателем или авто‑детект поступления)
- WAIT_PAYMENT → CANCELLED (отмена до оплаты: автор/продавец/система по `expiresAt`)
- PAID → RELEASED (подтверждение получения фиата продавцом или решением арбитража)
- PAID → DISPUTE (оспаривание любой стороной)
- DISPUTE → RELEASED (решение арбитража в пользу покупателя)
- DISPUTE → CANCELLED (решение арбитража в пользу продавца)

Терминальные статусы: RELEASED, CANCELLED.

## Роли и права

- WAIT_PAYMENT → PAID: покупатель (author) либо система (автодетект входящего платежа).
- WAIT_PAYMENT → CANCELLED: автор ордера; продавец — при отсутствии оплаты; система — по `expiresAt`.
- PAID → RELEASED: продавец (offerOwner) либо арбитраж.
- PAID → DISPUTE: любая сторона (author/offerOwner).
- DISPUTE → RELEASED/CANCELLED: только арбитраж (или системная роль).

## Поля модели (дополнения)

Рекомендуется расширить `models.Order` следующими полями (camelCase):

- `paidAt *time.Time` — момент подтверждения оплаты.
- `cancelReason *string` — причина отмены (если задана).
- `disputeReason *string` — причина открытия спора.
- `disputeOpenedAt *time.Time` — момент открытия спора.

Все новые поля возвращаются в JSON и попадают в Swagger (через теги).

## Эскроу и балансы

Если `order.isEscrow == true`:

- При создании: резервирование средств продавца в `AmountEscrow` по активу, который отдаёт продавец (как правило, crypto).
- При RELEASED: списать из эскроу у продавца и зачислить на баланс покупателя (создать `TransactionInternal` для аудита).
- При CANCELLED: вернуть зарезервированные средства из эскроу продавцу (также `TransactionInternal`).

Все операции выполняются атомарно в транзакции БД.

## Авто‑отмена по истечению

Фоновая задача (cron/воркер) периодически переводит ордера со статусом WAIT_PAYMENT в CANCELLED, если `expiresAt < now()`.
При этом создаются Notifications и рассылается событие в `/ws/orders/{id}/status`.

## События и уведомления

- WebSocket `/ws/orders/{id}/status`: при каждом изменении статуса шлётся `OrderStatusEvent { type: "order.status_changed", order: OrderFull }` автору и владельцу оффера.
- Notifications: для обеих сторон создаётся уведомление `order.status_changed` с payload `{ orderId, status }` и ссылкой на ордер.

## API‑эндпоинты статусов

Все ответы возвращают `OrderFull` в camelCase, включая связанные модели (Offer, Buyer, Seller, Author, OfferOwner, FromAsset, ToAsset, ClientPaymentMethod), как в текущем `GetOrder`.

1) Подтвердить оплату

- `POST /orders/{id}/paid`
- Body: `{ "paidAt"?: string(RFC3339) }` (если не передано — сохранить текущий момент)
- Права: только автор ордера (buyer)
- Допустимо из: WAIT_PAYMENT → PAID
- Эффекты: установить `paidAt`, создать notifications, WS‑событие `order.status_changed`

2) Выпустить средства (завершить)

- `POST /orders/{id}/release`
- Body: пустой
- Права: только продавец (offerOwner) или арбитраж
- Допустимо из: PAID → RELEASED
- Эффекты: провести движение эскроу → баланс покупателя, установить `releasedAt`, notifications, WS

3) Отменить

- `POST /orders/{id}/cancel`
- Body: `{ "reason"?: string }`
- Права: автор (до оплаты), продавец (при отсутствии оплаты), система (по истечению)
- Допустимо из: WAIT_PAYMENT → CANCELLED; из других статусов — только арбитраж
- Эффекты: вернуть эскроу продавцу (если был), notifications, WS

4) Открыть спор

- `POST /orders/{id}/dispute`
- Body: `{ "reason"?: string }`
- Права: любая сторона
- Допустимо из: PAID → DISPUTE
- Эффекты: установить `disputeReason`, `disputeOpenedAt`, notifications, WS

5) Решение спора

- `POST /orders/{id}/dispute/resolve`
- Body: `{ "result": "RELEASED"|"CANCELLED", "comment"?: string }`
- Права: арбитраж/система
- Допустимо из: DISPUTE → RELEASED|CANCELLED
- Эффекты: при RELEASED — перевод из эскроу покупателю; при CANCELLED — возврат эскроу продавцу; notifications, WS

## Идемпотентность и конкурентный доступ

- Все изменения статуса выполняются в транзакции с проверкой текущего статуса: `UPDATE orders SET status=?, ... WHERE id=? AND status=?`.
- При нулевом количестве обновлённых строк возвращаем 409 Conflict (или 400) с сообщением о некорректном переходе/гонке.
- Для денежных операций — отдельные записи `TransactionInternal` и, при необходимости, блокировка строк балансов.

## Пример событий WS `/ws/orders/{id}/status`

```json
{ "type": "order.status_changed", "order": { "id": "o123", "status": "PAID", "paidAt": "2025-01-01T00:00:00Z", "buyer": {"id":"...","username":"..."}, "seller": {"id":"...","username":"..."}, "offer": {"id":"..."}, "fromAsset": {"id":"..."}, "toAsset": {"id":"..."} } }
```

## Swagger (требования)

- Для каждого эндпоинта описать: Summary, Description (с правилами и допустимыми переходами), Security (BearerAuth), параметры/тела, коды ошибок, схемы ответов (`OrderFull`).
- Примеры запросов/ответов добавить для каждого перехода.
- Модель `Order` дополнить новыми полями (`paidAt`, `cancelReason`, `disputeReason`, `disputeOpenedAt`) и описать их.

## Тесты (обязательно)

Позитивные сценарии:

- WAIT_PAYMENT→PAID (по кнопке buyer и авто‑paidAt)
- PAID→RELEASED (перевод средств из эскроу, `releasedAt` установлен)
- PAID→DISPUTE (фиксируется причина/время)
- DISPUTE→RELEASED и DISPUTE→CANCELLED (решение арбитража)
- WAIT_PAYMENT→CANCELLED (ручная и авто по истечению)

Негативные сценарии:

- Запрещённые переходы (например, RELEASED→CANCELLED)
- Недостаточные права (buyer пытается release, seller — paid и т.п.)
- Гонки: двойной release/cancel — второе действие не меняет состояние

Сайд‑эффекты:

- WS `/ws/orders/{id}/status` — обе стороны получают `order.status_changed`
- Notifications созданы для buyer и seller (author/offerOwner)
- Эскроу/балансы скорректированы корректно на RELEASED/CANCELLED

## Совместимость и формат

- Все JSON поля — в camelCase.
- ID (если появляются новые сущности) — генерировать nanoid.
- Возвращать связанные модели в `OrderFull` (как в текущем API).

## План внедрения

1) Добавить поля модели + миграции
2) Реализовать эндпоинты (paid/release/cancel/dispute/resolve)
3) Добавить транзакционную бизнес‑логику и обновление эскроу/балансов
4) Уведомления и WS‑события
5) Swagger: схемы, описания, примеры
6) Тесты на хендлеры и побочные эффекты (включая WS)
7) Фоновая авто‑отмена по `expiresAt`


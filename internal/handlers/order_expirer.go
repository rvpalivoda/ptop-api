package handlers

import (
    "time"

    "gorm.io/gorm"

    "ptop/internal/models"
)

// OrderExpirer периодически отменяет просроченные ордера и рассылает события
type OrderExpirer struct {
    db       *gorm.DB
    interval time.Duration
    stopCh   chan struct{}
}

func NewOrderExpirer(db *gorm.DB, interval time.Duration) *OrderExpirer {
    if interval <= 0 {
        interval = 30 * time.Second
    }
    return &OrderExpirer{db: db, interval: interval, stopCh: make(chan struct{})}
}

// Start запускает периодическую проверку в отдельной горутине
func (e *OrderExpirer) Start() {
    ticker := time.NewTicker(e.interval)
    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                e.expireOnce()
            case <-e.stopCh:
                return
            }
        }
    }()
}

// Stop останавливает проверку
func (e *OrderExpirer) Stop() { close(e.stopCh) }

// expireOnce находит просроченные ордера и переводит их в CANCELLED,
// исключая ордера в статусе DISPUTE
func (e *OrderExpirer) expireOnce() {
    now := time.Now()
    var orders []models.Order
    if err := e.db.Where("expires_at <= ? AND status IN ?", now,
        []models.OrderStatus{models.OrderStatusWaitPayment, models.OrderStatusPaid}).
        Limit(100).Find(&orders).Error; err != nil {
        return
    }
    for _, ord := range orders {
        var res *gorm.DB
        // Для WAIT_PAYMENT -> CANCELLED, для PAID -> DISPUTE
        if ord.Status == models.OrderStatusWaitPayment {
            upd := map[string]any{"status": models.OrderStatusCancelled, "cancel_reason": "expired"}
            res = e.db.Model(&models.Order{}).
                Where("id = ? AND status = ?", ord.ID, models.OrderStatusWaitPayment).
                Updates(upd)
        } else if ord.Status == models.OrderStatusPaid {
            upd := map[string]any{"status": models.OrderStatusDispute, "dispute_opened_at": time.Now(), "dispute_reason": "expired"}
            res = e.db.Model(&models.Order{}).
                Where("id = ? AND status = ?", ord.ID, models.OrderStatusPaid).
                Updates(upd)
        } else {
            continue
        }
        if res.Error != nil || res.RowsAffected == 0 {
            continue
        }
        var full models.Order
        if err := e.db.Preload("Offer").
            Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
            Preload("FromAsset").Preload("ToAsset").
            Preload("ClientPaymentMethod").
            Preload("ClientPaymentMethod.Country").
            Preload("ClientPaymentMethod.PaymentMethod").
            Where("id = ?", ord.ID).First(&full).Error; err != nil {
            continue
        }
        CreateOrderStatusNotifications(e.db, full)
        BroadcastOrderStatus(full)
    }
}

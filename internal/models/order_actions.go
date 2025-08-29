package models

// OrderAction тип действия, доступного по ордеру
type OrderAction string

const (
    // OrderActionMarkPaid покупатель помечает ордер как оплаченный
    OrderActionMarkPaid OrderAction = "markPaid"
    // OrderActionCancel отмена ордера
    OrderActionCancel OrderAction = "cancel"
    // OrderActionDispute открытие спора
    OrderActionDispute OrderAction = "dispute"
    // OrderActionRelease продавец освобождает средства
    OrderActionRelease OrderAction = "release"
)


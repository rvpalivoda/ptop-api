package models

// OrderFull представляет ордер с вложенными данными связанных объектов
// swagger:model
// содержит ордер, оффер, покупателей/продавцов и связанные активы и метод оплаты

type OrderFull struct {
	Order
	Offer               Offer                `json:"offer"`
	Buyer               Client               `json:"buyer"`
	Seller              Client               `json:"seller"`
	FromAsset           Asset                `json:"fromAsset"`
	ToAsset             Asset                `json:"toAsset"`
	ClientPaymentMethod *ClientPaymentMethod `json:"clientPaymentMethod,omitempty"`
}

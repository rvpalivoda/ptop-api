package models

// OfferFull представляет объявление с вложенными данными связанных объектов
// swagger:model
// содержит оффер, активы, клиента и платёжные методы клиента

type OfferFull struct {
	Offer
	FromAsset            Asset                 `json:"fromAsset"`
	ToAsset              Asset                 `json:"toAsset"`
	Client               Client                `json:"client"`
	ClientPaymentMethods []ClientPaymentMethod `json:"clientPaymentMethods"`
	IsMine               bool                  `json:"isMine"`
}

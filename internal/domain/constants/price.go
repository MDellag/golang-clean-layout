package constants

type PriceType string

const (
	PriceTypeList  PriceType = "ListPrice"
	PriceTypeCost  PriceType = "CostPrice"
	PriceTypeSale  PriceType = "SalePrice"
	PriceTypeOffer PriceType = "OfferPrice"
)

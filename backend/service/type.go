package service

type bodyAsset struct{
	Name string `json:"name" form:"name" binding:"required"`
	Quantity float64  `json:"quantity" form:"quantity" binding:"required"`
}

type transition struct{
	UserId int
	AssetId int
	AsssetName string
	Price float64
	Quantity float64
}

type responseData struct{
	Price float64 `json:"p"`
	AssetSymbol string `json:"s"`
	Time int64 `json:"t"`
	Volume float64 `json:"v"`
}
type response struct{
	Data []responseData `json:"data"`
	Type string  `json:"type"`
}

type assetData struct{
	Id int 
	AsssetName string
	responseData
}

type assetSymbol string

type assetResponseData struct{
	AssetSymbol string `json:"s"`
	StockPriceNow float64 `json:"stock_price_now"`
	Income string  `json:"income"`
}
type assetResponse struct{
	Data []assetResponseData `json:"data"`
	Type string  `json:"type"`
}
package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	db "github/ukilolll/trade/database"
	"github/ukilolll/trade/pkg"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var(
	dbCon =db.Connect()
	Dashboard = make(map[assetSymbol]*assetData)//map use struct must be pointer
	responseStrcuture  []responseAssetData
	upgrader = &ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {return true},
	}

)

func showDashboad(name assetSymbol){
	data:= Dashboard[name] //map should cache data 
	fmt.Printf("%v price:%v time:%v\n",data.AssetSymbol,data.Price,time.Unix(data.Time,0).UTC())
}

func RunDashboard(){
	log.Println("running board")
	var res response
	url := fmt.Sprintf("wss://ws.finnhub.io?token=%v", os.Getenv("FINHUB_TOKEN") )
	conn,_,err := ws.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Panic(err)
	}
	rows,err := dbCon.Query("SELECT * FROM assets")
	if err != nil {
		log.Panic(err)
	}

	var i int
	for rows.Next(){
		var symbol assetSymbol
		var data assetData
		err := rows.Scan(&data.Id,&symbol,&data.AsssetName)
		if err != nil {
			log.Panic(err)
		}
		Dashboard[symbol]= &data
		i++;
	}
	rows.Close()

	//send data to query stock data
	for k,_ := range Dashboard{
		msg,_ := json.Marshal(map[string]any{"type": "subscribe", "symbol":k})
		log.Println(string(msg))
		err = conn.WriteMessage(ws.TextMessage,msg)
		if err != nil {
			log.Panicln(err)
		}
	}

	for{
		err := conn.ReadJSON(&res)
		if err != nil {
			break
		}	

		if res.Type == "trade" {


		for symbol,_ := range Dashboard{
		for _,resData := range res.Data{
			if string(symbol) == resData.AssetSymbol{
				Dashboard[symbol].responseData = resData
				//ShowDashboad(symbol)
			}
		// "asset_id", assetData.Id,"\n",
		// "asset_name", assetData.AsssetName,"\n",
		// "asset_symbol", symbol,"\n",
		// "price", assetData.Price,"\n",)
		}
		}
		var temp []responseAssetData
		for assetSymbol,assetdata := range Dashboard{
			temp = append(temp, responseAssetData{
				Id: assetdata.Id,
				AsssetName: assetdata.AsssetName,
				Price: assetdata.Price,
				Symbol: assetSymbol,
			})
		}
		responseStrcuture = temp

	}
	}

}

func Test(){
	fmt.Println(os.Getenv("FINHUB_TOKEN"))
}

func GetAssetDataThatHandle(ctx *gin.Context){
	var data []any
	for symbol,assetData := range Dashboard{
		data =append(data, map[string]any{
			"asset_id": assetData.Id,
			"asset_name": assetData.AsssetName,
			"asset_symbol": symbol,
			"price": assetData.Price,
		})
	}

	ctx.JSON(200,data)
}

func GetAssetDataRealTime(ctx *gin.Context){
	conn,err := upgrader.Upgrade(ctx.Writer,ctx.Request,nil)	
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(2 * time.Second)

	for{
	select{
	case _ = <-ticker.C:
		conn.WriteJSON(responseStrcuture)

	}
	}

}

func CheckAsset(ctx *gin.Context){
	var body bodyAsset
	var user_id,_ = strconv.Atoi(ctx.MustGet("id").(string))
	err := ctx.Bind(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest,gin.H{"msg":err.Error()})
// Key: 'bodyAsset.Name' Error:Field validation for 'Name' failed on the 'required' tag
// Key: 'bodyAsset.Quntity' Error:Field validation for 'Quntity' failed on the 'required' tag
		ctx.Abort();return;
	}

	if body.Quantity < 0 {
		pkg.BadRequest.SendErr("quntity should more than 0",ctx)
		ctx.Abort();return;
	}


	asset,ok := Dashboard[assetSymbol(body.Name)]
	if !ok{
		ctx.JSON(http.StatusBadRequest,gin.H{"msg":"invalid asset"})
		ctx.Abort();return;
	}

	transition := &transition{
		UserId: user_id,
		AssetId: asset.Id,
		AsssetName: body.Name,
		Quantity:body.Quantity ,
		Price: asset.Price,
	}

	ctx.Set("transition",transition)
	ctx.Next()
}

func minusUsercoin(trans *sql.Tx,userId int ,assetPrice float64 , quntity float64) error {
	var coin float64
	err := trans.QueryRow("SELECT coin FROM users WHERE user_id = $1",userId).Scan(&coin)
	if err != nil {
		return err
	}

	updateCoin := coin - ( assetPrice * float64(quntity) )
	if updateCoin < 0{
		return fmt.Errorf("coin not enough")
	}

	_,err = trans.Exec("UPDATE users SET coin = $1 WHERE user_id = $2;",updateCoin,userId)
	if err != nil {
		return err
	}

	return nil
}

func makeTransition(trans *sql.Tx , tradeType string ,data *transition ) error{
	command:=fmt.Sprintf(`INSERT INTO transition (trade_type, price, quantity, user_id, asset_id) 
	VALUES ('%v', $1, $2, $3, $4);`,
	tradeType)
	_, err := trans.Exec(command,data.Price, data.Quantity, data.UserId, data.AssetId)
	return err
}

func BuyAsset(ctx *gin.Context) {
	data , ok := ctx.MustGet("transition").(*transition)
	if !ok {
		pkg.Internal.SendErr("internal server error",ctx)
		return
	}	

	log.Println(data)

	trans ,err  := dbCon.Begin()
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk {
		log.Println("call")
		return
	}
	defer trans.Rollback()

	err = minusUsercoin(trans,data.UserId,data.Price,data.Quantity)
	if err != nil {
		if(err.Error() == "coin not enough"){
			ctx.JSON(http.StatusBadRequest,gin.H{"msg":"coin not enough"} )
			return
		}
		ctx.JSON(500,gin.H{"msg":"internal server error"})
		log.Println(err.Error())
		return
	}

	err = makeTransition(trans,"Buy",data)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	var oldPrice , oldQuantity float64 

	command := "SELECT quantity,price FROM user_assets WHERE user_id = $1 AND asset_id = $2;"
	err = trans.QueryRow(command,data.UserId,data.AssetId).Scan(&oldQuantity,&oldPrice)
	if (err != sql.ErrNoRows && err != nil){
		pkg.Internal.SendErr("internal server error",ctx)
		return
	}
	//first buy
	if(err == sql.ErrNoRows){
		_, err = trans.Exec("INSERT INTO user_assets(asset_id, user_id, quantity , price) VALUES($1, $2, $3, $4);",
		data.AssetId,data.UserId,data.Quantity,data.Price)
		if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
			return
		}
	}else{//not first buy
		newQuantity:= data.Quantity + oldQuantity
		newPrice := ( ( oldQuantity *oldPrice ) + ( data.Quantity * data.Price) ) / float64(newQuantity)

		_,err = trans.Exec("UPDATE user_assets SET quantity = $1 ,price = $2 WHERE user_id = $3 AND asset_id = $4;",
		newQuantity,newPrice,data.UserId,data.AssetId)
		if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
			return
		}
	}

	if err := trans.Commit(); err != nil {
		pkg.Internal.IsErr(err, "commit failed", ctx)
		return
	}
	ctx.JSON(200,gin.H{"msg": "buy success"})
}


func SellAsset(ctx *gin.Context)  {
	data , ok := ctx.MustGet("transition").(*transition)
	if !ok {
		pkg.Internal.SendErr("internal server error",ctx)
		return
	}	

	trans ,err  := dbCon.Begin()
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}
	defer trans.Rollback()

	err = makeTransition(trans,"Sell",data)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	var oldQuantity float64
	command := `
    SELECT quantity
    FROM user_assets
    WHERE user_id = $1 AND asset_id = $2
    FOR UPDATE`
	err = trans.QueryRow(command ,data.UserId,data.AssetId).Scan(&oldQuantity)
	log.Println(oldQuantity,data.Quantity)
	if data.Quantity > oldQuantity{
		ctx.JSON(http.StatusBadRequest,gin.H{"msg":"your request quantity amount more that your asset amount"} )
		return;
	}
	newQuantity := oldQuantity - data.Quantity

	log.Println("newQuantity:",newQuantity)
	if newQuantity <= 0 {
	//delete user asset
 	_, err = trans.Exec("DELETE FROM user_assets WHERE user_id = $1 AND asset_id = $2;",
	data.UserId,data.AssetId)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	}else{
	// update user asset
	command = `
	UPDATE user_assets 
	SET quantity = $1 WHERE user_id = $2 AND asset_id = $3;`
	_,err = trans.Exec(command,
	newQuantity,data.UserId,data.AssetId)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	}

	//update coin user
	var profit = data.Price * data.Quantity
	log.Println("profit:",profit)

	_,err =trans.Exec("UPDATE users SET coin = coin + $1 WHERE user_id = $2;",
	profit,data.UserId)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	if err := trans.Commit(); err != nil {
		pkg.Internal.IsErr(err, "commit failed", ctx)
		return
	}

	ctx.JSON(200,gin.H{"msg": "sell success"})
}


func LookAsset(ctx *gin.Context){
	var user_id,_ = strconv.Atoi(ctx.MustGet("id").(string))
	var res assetResponse
	var userQuantity,userPrice float64
	
	command := `SELECT symbol , quantity , price FROM user_assets 
	INNER JOIN assets ON user_assets.asset_id=assets.asset_id 
	WHERE user_id = $1;`
	rows ,err := dbCon.Query(command,user_id)
	defer rows.Close()
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}
	
	var firstTimeData []any
	for rows.Next(){
		var r assetResponseData
		rows.Scan(&r.AssetSymbol,&userQuantity,&userPrice)
		res.Data = append(res.Data, r)
		firstTimeData = append(firstTimeData, gin.H{
			"s": r.AssetSymbol,
			"quantity": userQuantity,
			"buy_price": userPrice,
		})
	}

	if(len(res.Data) == 0){
		ctx.JSON(http.StatusBadRequest,gin.H{"msg":"user have no asset now!"} )
		return
	}

	conn,err := upgrader.Upgrade(ctx.Writer,ctx.Request,nil)	
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}
	defer conn.Close()

	conn.WriteJSON(gin.H{
		"type": "not-real-time",
		"data": firstTimeData,
	})

	ticker := time.NewTicker(2 * time.Second)

	for{
	select{
	case _ = <-ticker.C:
		for i,v := range res.Data{
			data , ok := Dashboard[assetSymbol(v.AssetSymbol)]
			if !ok {
				conn.WriteJSON(gin.H{
					"type":  "error",
					"error": "you don't have asset now",
				})
				break
			}
			difference := data.Price / userPrice
			NowQuantity := userQuantity*difference
	
			var showPercent string
			percent := (difference-1)*100//difference 0.0 - infintie
			if percent > 0{
				showPercent = fmt.Sprintf("+%.6f",percent)
			}else{
				showPercent = fmt.Sprintf("%.6f",percent)
			}
	
			res.Data[i].Income = fmt.Sprintf("%.6f(%v%%)",NowQuantity,showPercent)
			res.Data[i].StockPriceNow = data.Price
			res.Type = "real-time"
		}
		conn.WriteJSON(res)
	}
	}	

}


func Reset(ctx *gin.Context){
	var user_id,_ = strconv.Atoi(ctx.MustGet("id").(string))
	trans,err := dbCon.Begin()
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}
	defer trans.Rollback()

	_, err = trans.Exec("UPDATE users SET coin = $1 WHERE user_id = $2;",10000,user_id)
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	_, err = trans.Exec("DELETE FROM user_assets WHERE user_id = $1")
	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
		return
	}

	if err := trans.Commit(); err != nil {
		pkg.Internal.IsErr(err, "commit failed", ctx)
		return
	}
	ctx.JSON(200,gin.H{"msg": "reset success"})
}


// func GetTradeHistory(ctx *gin.Context){
// 	var user_id,_ = strconv.Atoi(ctx.MustGet("id").(string))

// 	command := `SELECT trade_type, price, quantity, asset_name, transition.created_at
// 	FROM transition
// 	INNER JOIN assets ON transition.asset_id=assets.asset_id	
// 	WHERE user_id = $1
// 	ORDER BY transition.created_at DESC;`			
	
// 	rows,err := dbCon.Query(command,user_id)	

// 	if notOk := pkg.Internal.IsErr(err,"internal server error",ctx); notOk  {
// 		return
// 	}		
// 	defer rows.Close()

// 	var res []tradeHistory
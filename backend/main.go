package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github/ukilolll/trade/auth"
	"github/ukilolll/trade/pkg"
	"github/ukilolll/trade/service"
	"github/ukilolll/trade/test"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var(
	_ = pkg.LoadEnv()
)

func main() {
	runServer()
}

func runServer(){
	r := gin.Default()

	var AllowOrigins =strings.Split(os.Getenv("ORIGIN"),",")


	r.Use(cors.New(cors.Config{
		AllowOrigins:     AllowOrigins,
		AllowMethods:     []string{"GET","POST","PUT","DELETE","OPTIONS"},
		AllowHeaders:     []string{"Origin","Content-Type","Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/service",service.GetAssetDataThatHandle)

	authR := r.Group("/auth")
	authR.GET("/google/login",auth.HandleGoogleLogin)
	authR.GET("/google/callback",auth.HandleGoogleCallback)
	authR.POST("/logout",auth.HandleLogout)
	authR.GET("/user/profile",auth.AuthMiddleware,auth.CheckAuth)

	trade := r.Group("/trade",auth.AuthMiddleware)
	trade.POST("/buy",service.CheckAsset,service.BuyAsset)
	trade.POST("/sell",service.CheckAsset,service.SellAsset)
	trade.GET("/check/assets",service.LookAsset)
	trade.DELETE("/reset",service.Reset)
 
	go service.RunDashboard()

	srv := &http.Server{
		Addr: fmt.Sprintf(":%v",os.Getenv("SERVER_PORT")),
		Handler: r,
		ReadTimeout: 2*time.Second,
		WriteTimeout: 2*time.Second,
		MaxHeaderBytes: 1<<20,
	}


	if err := srv.ListenAndServe(); err != nil{
		log.Panic(err)
	}
}

func runTest(){
	test.Test0()
}
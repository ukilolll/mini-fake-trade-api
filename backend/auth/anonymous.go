package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AnnoymousLogin(ctx *gin.Context) {
	email, err := uuid.NewUUID()
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	trans , err := dbCon.Begin()
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	defer trans.Rollback()


	var userId int
	log.Println(email)
	command := "INSERT INTO users(email, auth_host , coin) VALUES($1, $2 ,$3) RETURNING user_id;"
	err = trans.QueryRow(command, email, "annoymous", 10000).Scan(&userId)
	if err != nil {
		log.Panic(err)
	}
	
	
	log.Println("User ID:", userId, "email:", email)

	strToken, err := generateJWT(fmt.Sprintf("%v", userId), email.String())
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	trans.Commit()
	ctx.SetCookie(authCookie.name, strToken, int(authCookie.time), "/", "localhost", false, true)

	ctx.JSON(200,gin.H{"messsage":"login success"})
}
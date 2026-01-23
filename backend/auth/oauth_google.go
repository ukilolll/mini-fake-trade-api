package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	db "github/ukilolll/trade/database"
	"github/ukilolll/trade/pkg"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// not assign
// loadEnv not actually load environment in process
var (
	_ = pkg.LoadEnv()
)

// assing var
var (
	oauthStateString     = os.Getenv("OAUTHSTATE_STRING")
	OAUTH2_CLIENT_ID     = os.Getenv("OAUTH2_CLIENT_ID")
	OAUTH2_CLIENT_SECRET = os.Getenv("OAUTH2_CLIENT_SECRET")

	googleOAuthConfig = &oauth2.Config{
		ClientID:     OAUTH2_CLIENT_ID,
		ClientSecret: OAUTH2_CLIENT_SECRET,
		RedirectURL:  fmt.Sprintf("http://%v:%v/auth/google/callback", os.Getenv("SERVER_DOMAIN_NAME"), os.Getenv("SERVER_PORT")),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
	authCookie = &Cookie{
		name: "trade_auth_token",
		time: 14 * 24 * time.Hour,
	}

	dbCon = db.Connect()
)

func generateJWT(id string, username string) (string, error) {
	// log.Println(id, username)
	claims := &Claims{
		Id:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authCookie.time)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET) //encode to string
}
func validateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWT_SECRET, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AuthMiddleware(ctx *gin.Context) {
	strToken, err := ctx.Cookie(authCookie.name)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "cookie not found")
		ctx.Abort()
		return
	}
	claims, err := validateJWT(strToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err.Error())
		ctx.Abort()
		return
	}

	ctx.Set("id", claims.Id)
	ctx.Set("username", claims.Username)
	ctx.Next()
}

// google oauth2
func HandleMain(ctx *gin.Context) {
	html := `<html><body><a href="/auth/google/login">Login with Google</a></body></html>`
	fmt.Fprint(ctx.Writer, html)
}

// server redirect user to google for login with google
// then  redreict to callback route with code and state
// state and code is JUST QUERY PARAM
func HandleGoogleLogin(ctx *gin.Context) {
	//make url for redirect
	url := googleOAuthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	//redirect user to google for login
	http.Redirect(ctx.Writer, ctx.Request, url, http.StatusTemporaryRedirect)
}

// check state to protect csrf
// exchange code to geting token with google
// use token to geting user info
func HandleGoogleCallback(ctx *gin.Context) {
	if ctx.Request.FormValue("state") != oauthStateString {
		http.Error(ctx.Writer, "Invalid OAuth state", http.StatusUnauthorized)
		return
	}
	//get query name code(?code=...)
	code := ctx.Request.FormValue("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(ctx.Writer, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// use token to geting user info
	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(ctx.Writer, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(ctx.Writer, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// {
	// 	"email": "ukikillme@gmail.com",
	// 	"family_name": "อวิรุทธ์ชีวิน",
	// 	"given_name": "พีรวิชญ์",
	// 	"id": "110939356007572036285",
	// 	"name": "พีรวิชญ์ อวิรุทธ์ชีวิน",
	// 	"picture": "https://lh3.googleusercontent.com/a/ACg8ocK9M3_4TmDd9ShzkO2zGDFKtjUu2Etc7CC3b0mAf85N9MMnolRg=s96-c",
	// 	"verified_email": true
	//   }
	// userId := userInfo["id"].(string)
	email := userInfo["email"].(string)

	var userId int
	err = dbCon.QueryRow("SELECT user_id FROM users  WHERE email = $1 ;", email).Scan(&userId)
	if err != nil {
		log.Println(email, userId)
		if err == sql.ErrNoRows {
			command := "INSERT INTO users(email, auth_host,coin) VALUES($1, $2 ,$3) RETURNING user_id;"
			err = dbCon.QueryRow(command, email, "google", 10000).Scan(&userId)
			if err != nil {
				log.Panic(err)
			}
		}
	}
	log.Println("User ID:", userId , "email:", email)

	strToken, err := generateJWT(fmt.Sprintf("%v", userId), email)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.SetCookie(authCookie.name, strToken, int(authCookie.time), "/", "localhost", false, true)
	//path
	// Cookie นี้จะถูกส่งไปเฉพาะหน้าเว็บที่อยู่ภายใต้ /admin เช่น
	// https://example.com/admin/dashboard
	// แต่ จะไม่ถูกส่ง ไปที่
	// https://example.com/profile
	//domain
	//กำหนดว่าคุกกี้จะถูกส่งไปยังโดเมนใด และสามารถใช้งานใน subdomain ได้

	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%v", os.Getenv("ORIGIN")))
}

func CheckAuth(ctx *gin.Context) {
	username, _ := ctx.Get("username")
	userId, _ := strconv.Atoi(ctx.MustGet("id").(string))
	log.Println(username , userId)

	var coin float64
	err := dbCon.QueryRow("SELECT coin FROM users WHERE user_id = $1", userId).Scan(&coin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "internal server error")
		log.Println(err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"username": username, "coin": coin})
}



func HandleLogout(ctx *gin.Context) {
	ctx.SetCookie(authCookie.name, "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

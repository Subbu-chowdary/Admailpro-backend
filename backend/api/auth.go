package api

import (
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/utils"
	"encoding/json"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func AuthMiddleware(next func(*fasthttp.RequestCtx)) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		authHeader := string(ctx.Request.Header.Peek("Authorization"))
		if authHeader == "" {
			ctx.Error("Unauthorized: No token provided", fasthttp.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			ctx.Error("Unauthorized: Invalid token format", fasthttp.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimSpace(authHeader[7:])
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.Error("Unauthorized: Invalid or expired token", fasthttp.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx.SetUserValue("email", claims.Email)
			next(ctx)
		} else {
			ctx.Error("Unauthorized: Invalid claims", fasthttp.StatusUnauthorized)
		}
	}
}

func SignupHandler(ctx *fasthttp.RequestCtx) {
	var u models.User
	if err := json.Unmarshal(ctx.PostBody(), &u); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	u.Password = string(hashed)
	u.ID = utils.GenerateID()
	if err := db.CreateUser(&u); err != nil {
		ctx.Error("Failed to create user", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
}

func LoginHandler(ctx *fasthttp.RequestCtx) {
	var u models.User
	if err := json.Unmarshal(ctx.PostBody(), &u); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	storedUser, err := db.FindUser(u.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(u.Password)) != nil {
		ctx.Error("Invalid credentials", fasthttp.StatusUnauthorized)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Email: u.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	})
	tokenString, _ := token.SignedString([]byte(config.GetConfig().JWTSecret))
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{"token": tokenString})
}

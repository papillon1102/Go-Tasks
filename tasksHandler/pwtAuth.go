package handler

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/papillon1102/go-tasks/models"
	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type AuthHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type PWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(collection *mongo.Collection, ctx context.Context, redis *redis.Client) *AuthHandler {
	return &AuthHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redis,
	}
}

var key = RandomString(32)

// Lock-Write implement for "id" (FIXME)
var id string

// PWT Middleware (NOTE)
func (ah *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenName := "token" + "_" + id
		tokenValue, err := ah.redisClient.Get(tokenName).Result()
		if err != nil {
			log.Error().Err(err)
			c.JSON(http.StatusForbidden, gin.H{"message": "not logged"})
			c.Abort()
		}

		// // Get token value from header
		// tokenValue := c.GetHeader("Authorization")

		// Create paseto
		paseto, err := NewPasetoMaker(key)
		if err != nil {
			log.Error().Err(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Verify the token
		_, err = paseto.DecryptPWT(tokenValue)
		if err != nil {
			log.Error().Err(err).Msg("Err verify token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

func (au *AuthHandler) SignInPWTHandler(c *gin.Context) {
	var authUser models.AuthUser
	if err := c.ShouldBindJSON(&authUser); err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cur := au.collection.FindOne(au.ctx, bson.M{
		"username": authUser.Username,
		"ggid":     authUser.Ggid,
	})

	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Create paseto
	paseto, err := NewPasetoMaker(key)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Make exp time
	expirationTime := time.Now().Add(time.Minute * 10)

	// Create paseto token
	footer := os.Getenv("JWT_SECRET")
	token, err := paseto.CreatePWT(authUser.Username, footer, expirationTime)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	id = authUser.Ggid

	// Add new session after user-signin with google-uuid

	saveTokenName := "token" + "_" + authUser.Ggid
	au.redisClient.Set(saveTokenName, token, 0)

	if err != nil {
		log.Debug().Err(err).Msg("Err in save sessions")
		return
	}

	log.Info().Msgf("Pwt %v\n", token)
	c.JSON(http.StatusOK, gin.H{"message": "Welcome back " + authUser.Username})
}

func (ah *AuthHandler) RefreshHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")

	paseto, err := NewPasetoMaker(key)
	if err != nil {
		log.Error().Err(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pwtToken, err := paseto.DecryptPWT(token)
	if err != nil {
		log.Error().Err(err).Msg("Err verify token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Check if token still has > 30 => not expired
	if time.Unix(pwtToken.Expiration.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is not expired"})
		return
	}

	expTime := time.Now().Add(time.Minute * 5)
	pwtToken.Expiration = expTime

	// Do we need to create new Paseto ? (FIXME)
	footer := os.Getenv("JWT_SECRET")
	newToken, err := paseto.CreatePWT(pwtToken.Issuer, footer, expTime)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	log.Info().Msgf("Pwt %v\n", newToken)
	c.JSON(http.StatusOK, token)

}

// JWT Middleware (FIXME)
// func (ah *AuthHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		// Get token value from header
// 		tokenValue := c.GetHeader("Authorization")
// 		claims := &Claims{}

// 		// Parse token value
// 		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte(os.Getenv("JWT_SECRET")), nil
// 		})

// 		if err != nil {
// 			log.Debug().Msgf("err", err)
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		// if token not valid => return unauthor status
// 		if !tkn.Valid {
// 			log.Debug().Msg("token validerr")
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		c.Next()
// 	}
// }

// JWT Signin-Handler (FIXME)
// func (au *AuthHandler) SignInHandler(c *gin.Context) {
// 	var authUser models.AuthUser
// 	if err := c.ShouldBindJSON(&authUser); err != nil {
// 		log.Error().Err(err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if authUser.Username != "admin" || authUser.Password != "password" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
// 		return
// 	}

// 	// Create expiration time & claims
// 	expirationTime := time.Now().Add(time.Minute * 10)
// 	claims := &Claims{
// 		Username: authUser.Username,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	// Create new token
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError,
// 			gin.H{"error": err.Error()})
// 		return
// 	}
// 	jwtOutput := PWTOutput{
// 		Token:   tokenString,
// 		Expires: expirationTime,
// 	}
// 	c.JSON(http.StatusOK, jwtOutput)
// }

package handler

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/papillon1102/Go-Tasks/models"
	"gopkg.in/mgo.v2/bson"

	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
	id          string
	locker      *sync.Mutex // Former: sync.Muutex
}

type PWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(collection *mongo.Collection, ctx context.Context, redis *redis.Client, id string, mutex *sync.Mutex) *AuthHandler { // former: sync.Mutex
	return &AuthHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redis,
		id:          id,
		locker:      mutex,
	}
}

var key = RandomString(32)

// Lock-Write implement for "id" (FIXME)
var id string

// PWT Middleware (NOTE)
func (ah *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenName := "token" + "_" + ah.id

		_, err := ah.redisClient.Get(tokenName).Result()
		if err != nil {
			log.Error().Err(err).Msg("Err from middleware")
			c.JSON(http.StatusForbidden, gin.H{"message": "not logged"})
			c.Abort()
		}

		// // Get token value from header
		// tokenValue := c.GetHeader("Authorization")

		// Create paseto <- don't need to (FIXME)
		// paseto, err := NewPasetoMaker(key)
		// if err != nil {
		// 	log.Error().Err(err)
		// 	c.AbortWithStatus(http.StatusBadRequest)
		// 	return
		// }

		// Verify the token <- maybe don't need (FIXME)
		// _, err = paseto.DecryptPWT(tokenValue)
		// if err != nil {
		// 	log.Error().Err(err).Msg("Err verify token")
		// 	c.AbortWithStatus(http.StatusUnauthorized)
		// 	return
		// }

		c.Next()
	}
}

// User-signin
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
	footer := os.Getenv("FOOTER")
	token, err := paseto.CreatePWT(authUser.Username, footer, expirationTime)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	// We need to lock here to ensure data written correctly
	au.locker.Lock()
	au.id = authUser.Ggid
	au.locker.Unlock()

	// Add new session after user-signin with google-uuid
	saveTokenName := "token" + "_" + authUser.Ggid

	// We will use mongoDB as session
	// management tools instead Redis (FIXME)
	au.redisClient.Set(saveTokenName, token, 0)

	if err != nil {
		log.Debug().Err(err).Msg("Err in save sessions")
		return
	}

	log.Info().Msgf("Pwt %v\n", token)
	c.JSON(http.StatusOK, gin.H{"message": "Welcome back " + authUser.Username})
}

func (au *AuthHandler) SignUpPWTHandler(c *gin.Context) {
	var authUser models.AuthUser
	if err := c.ShouldBindJSON(&authUser); err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := au.collection.InsertOne(au.ctx, authUser)
	if err != nil {
		log.Error().Err(err).Msg("Err from insert new user")
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		c.Abort()
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been added"})
}

// Renew-token for user
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

package handler

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/boj/redistore"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"
	"github.com/phuslu/log"
)

type Session struct {
	Value string `json:"value"`
}

var (
	Pool               *redis.Pool
	Store              *redistore.RediStore
	domainName         string
	sessionStoreSecret string
	sessionTimeOut     int
)

type SessionObj struct {
	Session map[string]interface{}
}

// Check again -> Should we use it (FIXME)
func (au *AuthHandler) AddSession(c *gin.Context) {

	var sess Session

	if err := c.ShouldBindJSON(&sess); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err := au.collection.InsertOne(au.ctx, sess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Err while add session"})
		return
	}

}

func newPool(server string) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Create New SessionStore
func InitSessionStore(redisServer, redisPort, domainName, sessionStoreSecret string) {

	Pool = newPool(redisServer + ":" + redisPort)
	sessionTimeOut = 300

	redisStore, err := redistore.NewRediStoreWithPool(Pool, []byte(sessionStoreSecret))
	redisStore.DefaultMaxAge = sessionTimeOut
	if err != nil {
		log.Error().Err(err).Msg("err creating new redis store")
		return
	}

	Store = redisStore

	log.Info().Msg("Registering session object")
	gob.Register(&SessionObj{})

	domainName = domainName
	sessionStoreSecret = sessionStoreSecret
}

// Get session from redis
func GetSession(r *http.Request, sessName, sessKey string) (sessObj *SessionObj, err error) {

	// Get a session
	sess, err := Store.Get(r, sessName)
	if err != nil {
		log.Error().Err(err).Msg("Err getting session")
		return
	}

	if obj, ok := sess.Values[sessKey].(*SessionObj); !ok {
		log.Error().Err(err).Msgf("Session not found for key %s", sessKey)
		err = fmt.Errorf("Session not found for key %s", sessKey)

	} else {
		sessObj = obj
	}

	return sessObj, err
}

func SaveSessionToStore(r *http.Request, w http.ResponseWriter, sessName, sessKey string, timeout int, sessionObj *SessionObj) {

	sess, err := Store.Get(r, sessName)
	if err != nil {
		log.Error().Err(err).Msg("Err getting session")
	}

	// Get Session-Object from sessKey
	sess.Values[sessKey] = sessionObj

	if sessionObj != nil {
		log.Info().Msgf("New session created for key %s", sessKey)
	}

	sess.Options = &sessions.Options{
		Domain:   domainName,
		Path:     "/",
		MaxAge:   timeout,
		HttpOnly: true,
	}

	if err = sess.Save(r, w); err != nil {
		log.Fatal().Msgf("Err saving session: %v\n", err)
	}

}

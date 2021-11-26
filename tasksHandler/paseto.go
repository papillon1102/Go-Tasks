package handler

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/o1egl/paseto"
	"github.com/phuslu/log"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

type Payload struct {
	ID        string
	Username  string
	IssuedAt  time.Time
	ExpiredAt time.Time
}

func NewPasetoMaker(symmetricKey string) (*PasetoMaker, error) {

	// if len(symmetricKey) != chachapoly1305.KeySize {
	// 	return nil, fmt.Errorf("invalid key size")
	// }

	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	return &Payload{
		ID:        "7ybhsdfb1934",
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}

func (p *Payload) Valid() error {

	// If current time is after exp time
	if time.Now().After(p.ExpiredAt) {
		return fmt.Errorf("token expired")
	}

	return nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {

	// Make new payload
	payload, _ := NewPayload(username, duration)

	// Encrypt
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {

	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		log.Error().Err(err).Msg("Invalid token")
		return nil, err
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

var alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// Paseto V1 Creator
func (maker *PasetoMaker) CreatePWT(username string, foot string, expTime time.Time) (eToken string, err error) {

	now := time.Now()
	exp := now.Add(4 * time.Minute)
	nbt := now
	footer := foot
	jsonToken := paseto.JSONToken{
		Audience:   "test",
		Issuer:     username,
		Jti:        "123",
		Subject:    "test_subject",
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  nbt,
	}

	//! Consider using footer as the answer for secret question .
	//! iF footer = "", => could cause error.
	eToken, err = paseto.NewV1().Encrypt(maker.symmetricKey, jsonToken, footer)
	if err != nil {

		log.Error().Err(err)
		return
	} else {
		return eToken, nil
	}
}

// Paseto V1 Decrypt
func (maker *PasetoMaker) DecryptPWT(token string) (pwtToken paseto.JSONToken, err error) {
	var newJsonToken paseto.JSONToken
	var newFooter string
	err = paseto.NewV1().Decrypt(token, maker.symmetricKey, &newJsonToken, &newFooter)
	if err != nil {
		return pwtToken, err
	} else {
		pwtToken = newJsonToken
		return pwtToken, nil
	}
}

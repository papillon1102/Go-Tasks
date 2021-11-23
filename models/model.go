package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Name  string `json:"name"`
	GGID  string `json:"ggid"`
	Email string `json:"email"`
}

type Task struct {
	// We need to add bson & json to avoid have 2 ID
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson: "name"`
	Tags      []string           `json:"tags" bson: "tags"`
	User      User               `json:"user" bson: "user"`
	Status    string             `json:"status" bson: "status"`
	CreatedAt time.Time          `json:"createdat" bson:"createat"`
}

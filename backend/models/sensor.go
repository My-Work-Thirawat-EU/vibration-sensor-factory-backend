package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sensor struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Location string             `json:"location" bson:"location"`
	Picture  string             `json:"picture" bson:"picture"` // This will store the URL or base64 of the picture
}

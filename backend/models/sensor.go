package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type Sensor struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Coordinates Coordinates        `json:"coordinates" bson:"coordinates"`
	Picture     string             `json:"picture" bson:"picture"` // This will store the URL or base64 of the picture
}

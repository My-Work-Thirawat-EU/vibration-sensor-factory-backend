package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sensor struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	SerialNumber string             `json:"serial_number" bson:"serial_number"`
	APIKey       string             `json:"api_key" bson:"api_key"`
	Location     string             `json:"location" bson:"location"`
	Picture      string             `json:"picture" bson:"picture"`
	FMax         float64            `json:"fmax" bson:"fmax"`
	LOR          float64            `json:"lor" bson:"lor"`
	GMax         float64            `json:"g_max" bson:"g_max"`
	AlarmThs     float64            `json:"alarm_ths" bson:"alarm_ths"`
	Token        string             `json:"token,omitempty" bson:"token,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}

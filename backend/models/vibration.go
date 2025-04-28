package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VibrationData struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SensorID  primitive.ObjectID `bson:"sensor_id" json:"sensor_id"`
	WarnID    primitive.ObjectID `bson:"warn_id" json:"warn_id"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	VX        string             `bson:"vx" json:"vx"` // string format เช่น base64, json, csv
	VY        string             `bson:"vy" json:"vy"`
	VZ        string             `bson:"vz" json:"vz"`
}

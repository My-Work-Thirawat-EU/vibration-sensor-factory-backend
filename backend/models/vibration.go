package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VibrationData struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`

	// FFT data for each axis
	FFTX []float64 `bson:"fft_x" json:"fft_x"`
	FFTY []float64 `bson:"fft_y" json:"fft_y"`
	FFTZ []float64 `bson:"fft_z" json:"fft_z"`

	// RMS values for each axis
	RMSX float64 `bson:"rms_x" json:"rms_x"`
	RMSY float64 `bson:"rms_y" json:"rms_y"`
	RMSZ float64 `bson:"rms_z" json:"rms_z"`

	// Peak values for each axis
	PeakX float64 `bson:"peak_x" json:"peak_x"`
	PeakY float64 `bson:"peak_y" json:"peak_y"`
	PeakZ float64 `bson:"peak_z" json:"peak_z"`
}

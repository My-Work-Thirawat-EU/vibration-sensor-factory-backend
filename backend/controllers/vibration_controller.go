package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateVibration(c *gin.Context) {
	var vibration models.VibrationData
	if err := c.ShouldBindJSON(&vibration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate serial number
	if vibration.SerialNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Serial number is required"})
		return
	}

	// Check if sensor exists
	sensorCollection := config.GetCollection("sensors")
	var sensor models.Sensor
	err := sensorCollection.FindOne(context.Background(), bson.M{"serial_number": vibration.SerialNumber}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid serial number"})
		return
	}

	// Validate FFT data
	if len(vibration.FFTX) == 0 || len(vibration.FFTY) == 0 || len(vibration.FFTZ) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FFT data is required for all axes"})
		return
	}

	vibration.CreatedAt = time.Now()

	collection := config.GetCollection("vibrations")
	result, err := collection.InsertOne(context.Background(), vibration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}

	vibration.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, vibration)
}

func GetVibrations(c *gin.Context) {
	var vibrations []models.VibrationData
	collection := config.GetCollection("vibrations")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip := (page - 1) * limit

	filter := bson.M{}

	if serialNumber := c.Query("serial_number"); serialNumber != "" {
		filter["serial_number"] = serialNumber
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter["created_at"] = bson.M{"$gte": t}
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			if _, ok := filter["created_at"]; ok {
				filter["created_at"].(bson.M)["$lte"] = t
			} else {
				filter["created_at"] = bson.M{"$lte": t}
			}
		}
	}

	// Add pagination options
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var vib models.VibrationData
		cursor.Decode(&vib)
		vibrations = append(vibrations, vib)
	}

	// Get total count for pagination
	total, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": vibrations,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func GetVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var vib models.VibrationData
	collection := config.GetCollection("vibrations")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&vib)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, vib)
}

func UpdateVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var vib models.VibrationData
	if err := c.ShouldBindJSON(&vib); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("vibrations")
	update := bson.M{
		"$set": bson.M{
			"serial_number": vib.SerialNumber,
			"fft_x":         vib.FFTX,
			"fft_y":         vib.FFTY,
			"fft_z":         vib.FFTZ,
			"rms_x":         vib.RMSX,
			"rms_y":         vib.RMSY,
			"rms_z":         vib.RMSZ,
			"peak_x":        vib.PeakX,
			"peak_y":        vib.PeakY,
			"peak_z":        vib.PeakZ,
		},
	}

	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vibration data updated"})
}

func DeleteVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := config.GetCollection("vibrations")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vibration data deleted"})
}

func BatchRegisterVibrations(c *gin.Context) {
	var vibrations []models.VibrationData
	if err := c.ShouldBindJSON(&vibrations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate each vibration entry
	for _, vibration := range vibrations {
		// Validate serial number
		if vibration.SerialNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Serial number is required for all entries"})
			return
		}

		// Check if sensor exists
		sensorCollection := config.GetCollection("sensors")
		var sensor models.Sensor
		err := sensorCollection.FindOne(context.Background(), bson.M{"serial_number": vibration.SerialNumber}).Decode(&sensor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid serial number: " + vibration.SerialNumber})
			return
		}

		// Validate FFT data
		if len(vibration.FFTX) == 0 || len(vibration.FFTY) == 0 || len(vibration.FFTZ) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FFT data is required for all axes"})
			return
		}

		// Set created_at if not provided
		if vibration.CreatedAt.IsZero() {
			vibration.CreatedAt = time.Now()
		}
	}

	// Prepare documents for bulk insert
	documents := make([]interface{}, len(vibrations))
	for i, vibration := range vibrations {
		documents[i] = vibration
	}

	collection := config.GetCollection("vibrations")
	result, err := collection.InsertMany(context.Background(), documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Batch insert failed"})
		return
	}

	// Update IDs in the response
	for i, id := range result.InsertedIDs {
		vibrations[i].ID = id.(primitive.ObjectID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully registered batch of vibration data",
		"count":   len(vibrations),
		"data":    vibrations,
	})
}

// CreateVibrationWithAPIKey handles vibration data submission with API key authentication
func CreateVibrationWithAPIKey(c *gin.Context) {
	apiKey := c.Param("apikey")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required"})
		return
	}

	// Find sensor by API key
	sensorCollection := config.GetCollection("sensors")
	var sensor models.Sensor
	err := sensorCollection.FindOne(context.Background(), bson.M{"api_key": apiKey}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	var vibrationData models.VibrationData
	if err := c.ShouldBindJSON(&vibrationData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the provided serial number matches the sensor's serial number
	if vibrationData.SerialNumber != "" && vibrationData.SerialNumber != sensor.SerialNumber {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Serial number does not match the authenticated sensor"})
		return
	}

	// Set the serial number from the authenticated sensor
	vibrationData.SerialNumber = sensor.SerialNumber
	vibrationData.CreatedAt = time.Now()

	// Validate required fields
	if len(vibrationData.FFTX) == 0 || len(vibrationData.FFTY) == 0 || len(vibrationData.FFTZ) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FFT data is required for all axes"})
		return
	}

	// Insert the vibration data
	collection := config.GetCollection("vibrations")
	result, err := collection.InsertOne(context.Background(), vibrationData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store vibration data"})
		return
	}

	// Set the ID from the inserted document
	vibrationData.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, vibrationData)
}

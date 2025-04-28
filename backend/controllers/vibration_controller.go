package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateVibration(c *gin.Context) {
	var vibration models.VibrationData
	if err := c.ShouldBindJSON(&vibration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if vibration.Timestamp.IsZero() {
		vibration.Timestamp = time.Now()
	}

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

	filter := bson.M{}

	if sensorID := c.Query("sensor_id"); sensorID != "" {
		if id, err := primitive.ObjectIDFromHex(sensorID); err == nil {
			filter["sensor_id"] = id
		}
	}

	if warnID := c.Query("warn_id"); warnID != "" {
		if id, err := primitive.ObjectIDFromHex(warnID); err == nil {
			filter["warn_id"] = id
		}
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter["timestamp"] = bson.M{"$gte": t}
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			if _, ok := filter["timestamp"]; ok {
				filter["timestamp"].(bson.M)["$lte"] = t
			} else {
				filter["timestamp"] = bson.M{"$lte": t}
			}
		}
	}

	cursor, err := collection.Find(context.Background(), filter)
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

	c.JSON(http.StatusOK, vibrations)
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
			"sensor_id": vib.SensorID,
			"warn_id":   vib.WarnID,
			"timestamp": vib.Timestamp,
			"vx":        vib.VX,
			"vy":        vib.VY,
			"vz":        vib.VZ,
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

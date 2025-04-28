package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	result, err := collection.InsertOne(context.Background(), sensor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sensor.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, sensor)
}

func GetSensors(c *gin.Context) {
	var sensors []models.Sensor
	collection := config.GetCollection("sensors")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var sensor models.Sensor
		cursor.Decode(&sensor)
		sensors = append(sensors, sensor)
	}

	c.JSON(http.StatusOK, sensors)
}

func GetSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var sensor models.Sensor
	collection := config.GetCollection("sensors")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, sensor)
}

func UpdateSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	update := bson.M{
		"$set": bson.M{
			"user_id":       sensor.UserID,
			"serial_number": sensor.SerialNumber,
			"location":      sensor.Location,
			"picture":       sensor.Picture,
			"config": bson.M{
				"fmax":      sensor.Config.FMax,
				"lor":       sensor.Config.LOR,
				"g_max":     sensor.Config.GMax,
				"alarm_ths": sensor.Config.AlarmThs,
			},
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor updated successfully"})
}

func DeleteSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := config.GetCollection("sensors")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

func RegisterSensor(c *gin.Context) {
	var request struct {
		SerialNumber string `json:"serial_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if sensor exists
	collection := config.GetCollection("sensors")
	var sensor models.Sensor
	err := collection.FindOne(context.Background(), bson.M{"serial_number": request.SerialNumber}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sensor_id": sensor.ID.Hex(),
		"exp":       time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days expiration
	})

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte(config.GetConfig().JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"sensor_id": sensor.ID.Hex(),
	})
}

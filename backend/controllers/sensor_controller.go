package controllers

import (
	"context"
	"net/http"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate coordinates
	if sensor.Coordinates.Latitude < -90 || sensor.Coordinates.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude must be between -90 and 90 degrees"})
		return
	}
	if sensor.Coordinates.Longitude < -180 || sensor.Coordinates.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude must be between -180 and 180 degrees"})
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

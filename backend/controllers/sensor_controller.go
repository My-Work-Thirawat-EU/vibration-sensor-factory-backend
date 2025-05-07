package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// generateSecureKey generates a secure random key of specified length
func generateSecureKey(length int) (string, error) {
	keyBytes := make([]byte, length)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(keyBytes), nil
}

// generateSensorCredentials generates both token and API key for a sensor
func generateSensorCredentials() (string, string, error) {
	// Generate token (32 bytes = 64 hex characters)
	token, err := generateSecureKey(32)
	if err != nil {
		return "", "", err
	}

	// Generate API key (24 bytes = 48 hex characters)
	apiKey, err := generateSecureKey(24)
	if err != nil {
		return "", "", err
	}

	return token, apiKey, nil
}

// CreateSensor creates a new sensor
func CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set creation time
	sensor.CreatedAt = time.Now()

	// Validate user_id
	if sensor.UserID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Check if user exists
	userCollection := config.GetCollection("users")
	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": sensor.UserID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Check if serial number is unique
	sensorCollection := config.GetCollection("sensors")
	var existingSensor models.Sensor
	err = sensorCollection.FindOne(context.Background(), bson.M{"serial_number": sensor.SerialNumber}).Decode(&existingSensor)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serial number already exists"})
		return
	}

	// Insert sensor
	result, err := sensorCollection.InsertOne(context.Background(), sensor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sensor"})
		return
	}

	sensor.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, gin.H{
		"sensor": sensor,
	})
}

// GetSensors retrieves all sensors with optional filtering
func GetSensors(c *gin.Context) {
	var sensors []models.Sensor
	sensorCollection := config.GetCollection("sensors")

	// Build filter
	filter := bson.M{}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := primitive.ObjectIDFromHex(userID); err == nil {
			filter["user_id"] = id
		}
	}

	// Find sensors
	cursor, err := sensorCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sensors"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &sensors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode sensors"})
		return
	}

	c.JSON(http.StatusOK, sensors)
}

// GetSensor retrieves a sensor by ID
func GetSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	var sensor models.Sensor
	sensorCollection := config.GetCollection("sensors")
	err = sensorCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&sensor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sensor"})
		return
	}

	c.JSON(http.StatusOK, sensor)
}

// UpdateSensor updates a sensor by ID
func UpdateSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't allow updating certain fields
	sensor.ID = objectID
	sensor.CreatedAt = time.Time{} // Don't update creation time

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"user_id":       sensor.UserID,
			"serial_number": sensor.SerialNumber,
			"api_key":       sensor.APIKey,
			"location":      sensor.Location,
			"picture":       sensor.Picture,
			"fmax":          sensor.FMax,
			"lor":           sensor.LOR,
			"g_max":         sensor.GMax,
			"alarm_ths":     sensor.AlarmThs,
		},
	}

	sensorCollection := config.GetCollection("sensors")
	result, err := sensorCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sensor"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sensor updated successfully"})
}

// DeleteSensor deletes a sensor by ID
func DeleteSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	sensorCollection := config.GetCollection("sensors")
	result, err := sensorCollection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sensor"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sensor deleted successfully"})
}

// GetSensorBySerialNumber retrieves a sensor by serial number
func GetSensorBySerialNumber(c *gin.Context) {
	serialNumber := c.Param("serial_number")
	if serialNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serial number is required"})
		return
	}

	var sensor models.Sensor
	sensorCollection := config.GetCollection("sensors")
	err := sensorCollection.FindOne(context.Background(), bson.M{"serial_number": serialNumber}).Decode(&sensor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sensor"})
		return
	}

	c.JSON(http.StatusOK, sensor)
}

func generateTokenHex(length int) (string, error) {
	tokenBytes := make([]byte, length)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// BatchCreateSensors creates multiple sensors
func BatchCreateSensors(c *gin.Context) {
	var sensors []models.Sensor
	if err := c.ShouldBindJSON(&sensors); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	var results []models.Sensor
	var errors []string

	for _, sensor := range sensors {
		result, err := collection.InsertOne(context.Background(), sensor)
		if err != nil {
			errors = append(errors, "Error creating sensor: "+sensor.SerialNumber)
			continue
		}

		sensor.ID = result.InsertedID.(primitive.ObjectID)
		results = append(results, sensor)
	}

	response := gin.H{
		"successful_registrations": len(results),
		"failed_registrations":     len(errors),
		"sensors":                  results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if len(results) > 0 {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// RegisterSensor registers a sensor and generates credentials
func RegisterSensor(c *gin.Context) {
	var request struct {
		SerialNumber string `json:"serial_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	var sensor models.Sensor

	// Find sensor by serial number
	err := collection.FindOne(context.Background(), bson.M{"serial_number": request.SerialNumber}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	// Generate 32-byte token (64 hex characters)
	tokenString, err := generateTokenHex(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Generate API key (24 bytes = 48 hex characters)
	apiKey, err := generateTokenHex(24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating API key"})
		return
	}

	// Update sensor document with the generated credentials
	update := bson.M{
		"$set": bson.M{
			"token":   tokenString,
			"api_key": apiKey,
		},
	}

	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": sensor.ID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating sensor credentials"})
		return
	}

	// Extra safety: check if the sensor matched
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found during update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"api_key":   apiKey,
		"sensor_id": sensor.ID.Hex(),
	})
}

// generateToken generates a random token of specified length
func generateToken(length int) (string, error) {
	tokenBytes := make([]byte, length)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// GenerateSensorToken generates a new token for a sensor
func GenerateSensorToken(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	// Generate new token (32 bytes = 64 hex characters)
	token, err := generateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Update sensor with new token
	sensorCollection := config.GetCollection("sensors")
	update := bson.M{
		"$set": bson.M{
			"token": token,
		},
	}

	result, err := sensorCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sensor token"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// ValidateSensorToken validates a sensor's token
func ValidateSensorToken(c *gin.Context) {
	token := c.GetHeader("X-Sensor-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	var sensor models.Sensor
	sensorCollection := config.GetCollection("sensors")
	err := sensorCollection.FindOne(context.Background(), bson.M{
		"token": token,
	}).Decode(&sensor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"sensor_id": sensor.ID.Hex(),
	})
}

// RevokeSensorToken revokes a sensor's token
func RevokeSensorToken(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	sensorCollection := config.GetCollection("sensors")
	update := bson.M{
		"$unset": bson.M{
			"token": "",
		},
	}

	result, err := sensorCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
}

// RegenerateSensorCredentials generates new token and API key for a sensor
func RegenerateSensorCredentials(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	// Generate new credentials
	token, apiKey, err := generateSensorCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate credentials"})
		return
	}

	// Update sensor with new credentials
	sensorCollection := config.GetCollection("sensors")
	update := bson.M{
		"$set": bson.M{
			"token":   token,
			"api_key": apiKey,
		},
	}

	result, err := sensorCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sensor credentials"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credentials": gin.H{
			"token":   token,
			"api_key": apiKey,
		},
	})
}

// ValidateSensorCredentials validates both token and API key
func ValidateSensorCredentials(c *gin.Context) {
	token := c.GetHeader("X-Sensor-Token")
	apiKey := c.GetHeader("X-API-Key")

	if token == "" || apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "both token and API key are required"})
		return
	}

	var sensor models.Sensor
	sensorCollection := config.GetCollection("sensors")
	err := sensorCollection.FindOne(context.Background(), bson.M{
		"token":   token,
		"api_key": apiKey,
	}).Decode(&sensor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"sensor_id": sensor.ID.Hex(),
	})
}

// RevokeSensorCredentials revokes both token and API key
func RevokeSensorCredentials(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor id"})
		return
	}

	sensorCollection := config.GetCollection("sensors")
	update := bson.M{
		"$unset": bson.M{
			"token":   "",
			"api_key": "",
		},
	}

	result, err := sensorCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke credentials"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credentials revoked successfully"})
}

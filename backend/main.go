package main

import (
	"log"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize MongoDB connection
	err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	r := gin.Default()

	// Sensor routes
	r.POST("/sensors", controllers.CreateSensor)
	r.GET("/sensors", controllers.GetSensors)
	r.GET("/sensors/:id", controllers.GetSensor)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"Server": "Running"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.Run(":8080") // เปิดพอร์ต 8080

}

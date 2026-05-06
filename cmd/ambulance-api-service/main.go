package main

import (
    "log"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/HorvathDarius/dhk-ambulance-webapi/api"
	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/ambulance"
)

func main() {
    log.Printf("Server started")
    port := os.Getenv("AMBULANCE_API_PORT")
    if port == "" {
        port = "8080"
    }
    environment := os.Getenv("AMBULANCE_API_ENVIRONMENT")
    if !strings.EqualFold(environment, "production") { // case insensitive comparison
        gin.SetMode(gin.DebugMode)
    }
    engine := gin.New()
    engine.Use(gin.Recovery())
    // request routings
	handleFunctions := &ambulance.ApiHandleFunctions{
		PerformanceRecordsAPI:  ambulance.NewPerformanceRecordsApi(),
	}
	ambulance.NewRouterWithGinEngine(engine, *handleFunctions)
    engine.GET("/openapi", api.HandleOpenApi)
    engine.Run(":" + port)
}
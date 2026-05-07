package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/HorvathDarius/dhk-ambulance-webapi/api"
	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/ambulance"
	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/db_service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(corsMiddleware)

	// setup context update  middleware
	dbService := db_service.NewMongoService[ambulance.PerformanceRecord](db_service.MongoServiceConfig{})
	defer dbService.Disconnect(context.Background())

	assignmentCollection := os.Getenv("AMBULANCE_API_MONGODB_ASSIGNMENT_COLLECTION")
	if assignmentCollection == "" {
		assignmentCollection = "department_assignments"
	}
	assignmentDbService := db_service.NewMongoService[ambulance.DepartmentAssignment](db_service.MongoServiceConfig{
		Collection: assignmentCollection,
	})
	defer assignmentDbService.Disconnect(context.Background())

	employeeCollection := os.Getenv("AMBULANCE_API_MONGODB_EMPLOYEE_COLLECTION")
	if employeeCollection == "" {
		employeeCollection = "employees"
	}
	employeeDbService := db_service.NewMongoService[ambulance.EmployeeProfile](db_service.MongoServiceConfig{
		Collection: employeeCollection,
	})
	defer employeeDbService.Disconnect(context.Background())

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", dbService)
		ctx.Set("db_service_assignment", assignmentDbService)
		ctx.Set("db_service_employee", employeeDbService)
		ctx.Next()
	})

	// request routings
	handleFunctions := &ambulance.ApiHandleFunctions{
		PerformanceRecordsAPI:    ambulance.NewPerformanceRecordsApi(),
		DepartmentAssignmentsAPI: ambulance.NewDepartmentAssignmentsApi(),
		EmployeeProfilesAPI:      ambulance.NewEmployeeProfilesApi(),
	}
	ambulance.NewRouterWithGinEngine(engine, *handleFunctions)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}

package ambulance

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implPerformanceRecordsAPI struct {
}

func NewPerformanceRecordsApi() PerformanceRecordsAPI {
	return &implPerformanceRecordsAPI{}
}

func (o implPerformanceRecordsAPI) CreatePerformanceRecord(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implPerformanceRecordsAPI) DeletePerformanceRecord(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implPerformanceRecordsAPI) GetPerformanceRecord(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implPerformanceRecordsAPI) ListPerformanceRecords(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implPerformanceRecordsAPI) UpdatePerformanceRecord(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

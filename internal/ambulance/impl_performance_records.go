package ambulance

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

type implPerformanceRecordsAPI struct {
}

func NewPerformanceRecordsApi() PerformanceRecordsAPI {
	return &implPerformanceRecordsAPI{}
}

// dbFromContext extracts the typed DbService injected by the middleware in main.go.
// Returns the service and true on success, or writes a 500 response and returns false.
func dbFromContext(c *gin.Context) (db_service.DbService[PerformanceRecord], bool) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
			"error":   "db_service not found",
		})
		return nil, false
	}
	db, ok := value.(db_service.DbService[PerformanceRecord])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of required type",
			"error":   "cannot cast db_service context to db_service.DbService",
		})
		return nil, false
	}
	return db, true
}

// parseRecordId reads the :recordId path parameter as int64.
func parseRecordId(c *gin.Context) (int64, bool) {
	raw := c.Param("recordId")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "recordId must be a numeric identifier",
			"error":   err.Error(),
		})
		return 0, false
	}
	return id, true
}

// CreatePerformanceRecord — POST /api/performance-records
func (o *implPerformanceRecordsAPI) CreatePerformanceRecord(c *gin.Context) {
	db, ok := dbFromContext(c)
	if !ok {
		return
	}

	record := PerformanceRecord{}
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// id is server-assigned per the spec; use nanosecond timestamp for uniqueness.
	record.Id = time.Now().UnixNano()

	err := db.CreateDocument(c.Request.Context(), record.Id, &record)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, record)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Conflict",
			"message": "Record with this id already exists",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to create performance record in database",
			"error":   err.Error(),
		})
	}
}

// ListPerformanceRecords — GET /api/performance-records
func (o *implPerformanceRecordsAPI) ListPerformanceRecords(c *gin.Context) {
	db, ok := dbFromContext(c)
	if !ok {
		return
	}

	records, err := db.ListDocuments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to list performance records",
			"error":   err.Error(),
		})
		return
	}

	// Optional filtering by query params (employeeId, from, to).
	employeeId := c.Query("employeeId")
	from := c.Query("from")
	to := c.Query("to")

	out := make([]PerformanceRecord, 0, len(records))
	for _, r := range records {
		if r == nil {
			continue
		}
		if employeeId != "" && r.EmployeeId != employeeId {
			continue
		}
		if from != "" && r.Date < from {
			continue
		}
		if to != "" && r.Date > to {
			continue
		}
		out = append(out, *r)
	}
	c.JSON(http.StatusOK, out)
}

// GetPerformanceRecord — GET /api/performance-records/:recordId
func (o *implPerformanceRecordsAPI) GetPerformanceRecord(c *gin.Context) {
	db, ok := dbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseRecordId(c)
	if !ok {
		return
	}

	record, err := db.FindDocument(c.Request.Context(), id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, record)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Performance record not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load performance record",
			"error":   err.Error(),
		})
	}
}

// UpdatePerformanceRecord — PUT /api/performance-records/:recordId
func (o *implPerformanceRecordsAPI) UpdatePerformanceRecord(c *gin.Context) {
	db, ok := dbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseRecordId(c)
	if !ok {
		return
	}

	record := PerformanceRecord{}
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}
	// Path id wins over any id provided in the body (per spec).
	record.Id = id

	err := db.UpdateDocument(c.Request.Context(), id, &record)
	switch err {
	case nil:
		c.JSON(http.StatusOK, record)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Performance record not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to update performance record",
			"error":   err.Error(),
		})
	}
}

// DeletePerformanceRecord — DELETE /api/performance-records/:recordId
func (o *implPerformanceRecordsAPI) DeletePerformanceRecord(c *gin.Context) {
	db, ok := dbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseRecordId(c)
	if !ok {
		return
	}

	err := db.DeleteDocument(c.Request.Context(), id)
	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Performance record not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to delete performance record",
			"error":   err.Error(),
		})
	}
}

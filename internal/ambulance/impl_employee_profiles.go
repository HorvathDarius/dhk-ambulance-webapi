package ambulance

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

type implEmployeeProfilesAPI struct {
}

func NewEmployeeProfilesApi() EmployeeProfilesAPI {
	return &implEmployeeProfilesAPI{}
}

func employeeDbFromContext(c *gin.Context) (db_service.DbService[EmployeeProfile], bool) {
	value, exists := c.Get("db_service_employee")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service_employee not found",
			"error":   "db_service_employee not found",
		})
		return nil, false
	}
	db, ok := value.(db_service.DbService[EmployeeProfile])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service_employee context is not of required type",
			"error":   "cannot cast db_service_employee context to db_service.DbService",
		})
		return nil, false
	}
	return db, true
}

func parseEmployeeId(c *gin.Context) (int64, bool) {
	raw := c.Param("employeeId")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "employeeId must be a numeric identifier",
			"error":   err.Error(),
		})
		return 0, false
	}
	return id, true
}

func nextEmployeeId(ctx context.Context, db db_service.DbService[EmployeeProfile]) (int64, error) {
	profiles, err := db.ListDocuments(ctx)
	if err != nil {
		return 0, err
	}
	var maxId int64
	for _, profile := range profiles {
		if profile != nil && profile.Id > maxId {
			maxId = profile.Id
		}
	}
	return maxId + 1, nil
}

func containsFold(value string, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}

func prepareEmployeeForCreate(profile *EmployeeProfile, id int64) {
	now := time.Now().UTC()
	profile.Id = id
	if profile.Status == "" {
		profile.Status = ACTIVE
	}
	profile.CreatedAt = now
	profile.UpdatedAt = time.Time{}
	profile.ArchivedAt = time.Time{}
}

func prepareEmployeeForUpdate(profile *EmployeeProfile, existing *EmployeeProfile, id int64) {
	now := time.Now().UTC()
	profile.Id = id
	if profile.Status == "" {
		profile.Status = existing.Status
	}
	if profile.Status == "" {
		profile.Status = ACTIVE
	}
	profile.CreatedAt = existing.CreatedAt
	profile.UpdatedAt = now
	if profile.Status == ARCHIVED && profile.ArchivedAt.IsZero() {
		profile.ArchivedAt = existing.ArchivedAt
	}
}

// CreateEmployeeProfile — POST /api/employees
func (o *implEmployeeProfilesAPI) CreateEmployeeProfile(c *gin.Context) {
	db, ok := employeeDbFromContext(c)
	if !ok {
		return
	}

	profile := EmployeeProfile{}
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	id, err := nextEmployeeId(c.Request.Context(), db)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to allocate employee id",
			"error":   err.Error(),
		})
		return
	}
	prepareEmployeeForCreate(&profile, id)

	err = db.CreateDocument(c.Request.Context(), profile.Id, &profile)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, profile)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Conflict",
			"message": "Employee profile with this id already exists",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to create employee profile in database",
			"error":   err.Error(),
		})
	}
}

// ListEmployeeProfiles — GET /api/employees
func (o *implEmployeeProfilesAPI) ListEmployeeProfiles(c *gin.Context) {
	db, ok := employeeDbFromContext(c)
	if !ok {
		return
	}

	profiles, err := db.ListDocuments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to list employee profiles",
			"error":   err.Error(),
		})
		return
	}

	department := c.Query("department")
	specialization := c.Query("specialization")
	status := c.Query("status")

	out := make([]EmployeeProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile == nil {
			continue
		}
		if department != "" && !containsFold(profile.Department, department) {
			continue
		}
		if specialization != "" && !containsFold(profile.Specialization, specialization) {
			continue
		}
		if status != "" && string(profile.Status) != status {
			continue
		}
		out = append(out, *profile)
	}
	c.JSON(http.StatusOK, out)
}

// GetEmployeeProfile — GET /api/employees/:employeeId
func (o *implEmployeeProfilesAPI) GetEmployeeProfile(c *gin.Context) {
	db, ok := employeeDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseEmployeeId(c)
	if !ok {
		return
	}

	profile, err := db.FindDocument(c.Request.Context(), id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, profile)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Employee profile not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load employee profile",
			"error":   err.Error(),
		})
	}
}

// UpdateEmployeeProfile — PUT /api/employees/:employeeId
func (o *implEmployeeProfilesAPI) UpdateEmployeeProfile(c *gin.Context) {
	db, ok := employeeDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseEmployeeId(c)
	if !ok {
		return
	}

	existing, err := db.FindDocument(c.Request.Context(), id)
	switch err {
	case nil:
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Employee profile not found",
			"error":   err.Error(),
		})
		return
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load employee profile",
			"error":   err.Error(),
		})
		return
	}

	profile := EmployeeProfile{}
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}
	prepareEmployeeForUpdate(&profile, existing, id)

	err = db.UpdateDocument(c.Request.Context(), id, &profile)
	switch err {
	case nil:
		c.JSON(http.StatusOK, profile)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Employee profile not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to update employee profile",
			"error":   err.Error(),
		})
	}
}

// ArchiveEmployeeProfile — DELETE /api/employees/:employeeId
func (o *implEmployeeProfilesAPI) ArchiveEmployeeProfile(c *gin.Context) {
	db, ok := employeeDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseEmployeeId(c)
	if !ok {
		return
	}

	profile, err := db.FindDocument(c.Request.Context(), id)
	switch err {
	case nil:
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Employee profile not found",
			"error":   err.Error(),
		})
		return
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load employee profile",
			"error":   err.Error(),
		})
		return
	}

	now := time.Now().UTC()
	profile.Status = ARCHIVED
	profile.ArchivedAt = now
	profile.UpdatedAt = now

	err = db.UpdateDocument(c.Request.Context(), id, profile)
	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Employee profile not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to archive employee profile",
			"error":   err.Error(),
		})
	}
}

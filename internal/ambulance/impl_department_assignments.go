package ambulance

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

type implDepartmentAssignmentsAPI struct {
}

func NewDepartmentAssignmentsApi() DepartmentAssignmentsAPI {
	return &implDepartmentAssignmentsAPI{}
}

// assignmentDbFromContext extracts the typed DbService for DepartmentAssignment.
func assignmentDbFromContext(c *gin.Context) (db_service.DbService[DepartmentAssignment], bool) {
	value, exists := c.Get("db_service_assignment")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service_assignment not found",
			"error":   "db_service_assignment not found",
		})
		return nil, false
	}
	db, ok := value.(db_service.DbService[DepartmentAssignment])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service_assignment context is not of required type",
			"error":   "cannot cast db_service_assignment context to db_service.DbService",
		})
		return nil, false
	}
	return db, true
}

// parseAssignmentId reads the :assignmentId path parameter as int64.
func parseAssignmentId(c *gin.Context) (int64, bool) {
	raw := c.Param("assignmentId")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "assignmentId must be a numeric identifier",
			"error":   err.Error(),
		})
		return 0, false
	}
	return id, true
}

// CreateDepartmentAssignment — POST /api/department-assignments
func (o *implDepartmentAssignmentsAPI) CreateDepartmentAssignment(c *gin.Context) {
	db, ok := assignmentDbFromContext(c)
	if !ok {
		return
	}

	assignment := DepartmentAssignment{}
	if err := c.BindJSON(&assignment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	assignment.Id = time.Now().UnixNano()

	err := db.CreateDocument(c.Request.Context(), assignment.Id, &assignment)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, assignment)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Conflict",
			"message": "Assignment with this id already exists",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to create department assignment in database",
			"error":   err.Error(),
		})
	}
}

// ListDepartmentAssignments — GET /api/department-assignments
func (o *implDepartmentAssignmentsAPI) ListDepartmentAssignments(c *gin.Context) {
	db, ok := assignmentDbFromContext(c)
	if !ok {
		return
	}

	assignments, err := db.ListDocuments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to list department assignments",
			"error":   err.Error(),
		})
		return
	}

	employeeId := c.Query("employeeId")
	departmentId := c.Query("departmentId")
	activeOn := c.Query("activeOn")

	out := make([]DepartmentAssignment, 0, len(assignments))
	for _, a := range assignments {
		if a == nil {
			continue
		}
		if employeeId != "" && a.EmployeeId != employeeId {
			continue
		}
		if departmentId != "" && a.DepartmentId != departmentId {
			continue
		}
		if activeOn != "" {
			if a.FromDate != "" && a.FromDate > activeOn {
				continue
			}
			if a.ToDate != "" && a.ToDate < activeOn {
				continue
			}
		}
		out = append(out, *a)
	}
	c.JSON(http.StatusOK, out)
}

// GetDepartmentAssignment — GET /api/department-assignments/:assignmentId
func (o *implDepartmentAssignmentsAPI) GetDepartmentAssignment(c *gin.Context) {
	db, ok := assignmentDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseAssignmentId(c)
	if !ok {
		return
	}

	assignment, err := db.FindDocument(c.Request.Context(), id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, assignment)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Department assignment not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load department assignment",
			"error":   err.Error(),
		})
	}
}

// UpdateDepartmentAssignment — PUT /api/department-assignments/:assignmentId
func (o *implDepartmentAssignmentsAPI) UpdateDepartmentAssignment(c *gin.Context) {
	db, ok := assignmentDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseAssignmentId(c)
	if !ok {
		return
	}

	assignment := DepartmentAssignment{}
	if err := c.BindJSON(&assignment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}
	assignment.Id = id

	err := db.UpdateDocument(c.Request.Context(), id, &assignment)
	switch err {
	case nil:
		c.JSON(http.StatusOK, assignment)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Department assignment not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to update department assignment",
			"error":   err.Error(),
		})
	}
}

// DeleteDepartmentAssignment — DELETE /api/department-assignments/:assignmentId
func (o *implDepartmentAssignmentsAPI) DeleteDepartmentAssignment(c *gin.Context) {
	db, ok := assignmentDbFromContext(c)
	if !ok {
		return
	}
	id, ok := parseAssignmentId(c)
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
			"message": "Department assignment not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to delete department assignment",
			"error":   err.Error(),
		})
	}
}

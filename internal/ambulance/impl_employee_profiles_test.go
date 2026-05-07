package ambulance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HorvathDarius/dhk-ambulance-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type employeeDbServiceMock struct {
	mock.Mock
}

func (m *employeeDbServiceMock) CreateDocument(ctx context.Context, id any, document *EmployeeProfile) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *employeeDbServiceMock) FindDocument(ctx context.Context, id any) (*EmployeeProfile, error) {
	args := m.Called(ctx, id)
	doc, _ := args.Get(0).(*EmployeeProfile)
	return doc, args.Error(1)
}

func (m *employeeDbServiceMock) UpdateDocument(ctx context.Context, id any, document *EmployeeProfile) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *employeeDbServiceMock) DeleteDocument(ctx context.Context, id any) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *employeeDbServiceMock) ListDocuments(ctx context.Context) ([]*EmployeeProfile, error) {
	args := m.Called(ctx)
	docs, _ := args.Get(0).([]*EmployeeProfile)
	return docs, args.Error(1)
}

func (m *employeeDbServiceMock) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func newEmployeeTestContext(db *employeeDbServiceMock, req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("db_service_employee", db_service.DbService[EmployeeProfile](db))
	c.Request = req
	return c, rec
}

func TestEmployeeProfilesCreateAssignsSequentialId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := &employeeDbServiceMock{}
	db.On("ListDocuments", mock.Anything).Return([]*EmployeeProfile{{Id: 2}}, nil)
	db.On("CreateDocument", mock.Anything, int64(3), mock.AnythingOfType("*ambulance.EmployeeProfile")).Return(nil)

	body := `{"firstName":"Anna","lastName":"Nováková","birthDate":"1988-04-12","position":"Lekárka","specialization":"Urgentná medicína","qualification":"MUDr.","employmentStartDate":"2026-06-01","certificates":["ALS"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/employees", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, rec := newEmployeeTestContext(db, req)

	api := &implEmployeeProfilesAPI{}
	api.CreateEmployeeProfile(c)

	require.Equal(t, http.StatusCreated, rec.Code)
	var created EmployeeProfile
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.Equal(t, int64(3), created.Id)
	require.Equal(t, ACTIVE, created.Status)
	require.False(t, created.CreatedAt.IsZero())
	db.AssertExpectations(t)
}

func TestEmployeeProfilesListFiltersDepartmentAndSpecialization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := &employeeDbServiceMock{}
	db.On("ListDocuments", mock.Anything).Return([]*EmployeeProfile{
		{Id: 1, FirstName: "Anna", LastName: "Nováková", Department: "Urgentný príjem", Specialization: "Urgentná medicína", Status: ACTIVE},
		{Id: 2, FirstName: "Peter", LastName: "Kováč", Department: "Kardiológia", Specialization: "Kardiologická sestra", Status: ACTIVE},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/employees?department=urgent&specialization=medicína", nil)
	c, rec := newEmployeeTestContext(db, req)

	api := &implEmployeeProfilesAPI{}
	api.ListEmployeeProfiles(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var profiles []EmployeeProfile
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &profiles))
	require.Len(t, profiles, 1)
	require.Equal(t, int64(1), profiles[0].Id)
	db.AssertExpectations(t)
}

func TestEmployeeProfilesArchiveUpdatesProfileStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := &employeeDbServiceMock{}
	profile := &EmployeeProfile{Id: 7, FirstName: "Anna", LastName: "Nováková", Status: ACTIVE}
	db.On("FindDocument", mock.Anything, int64(7)).Return(profile, nil)
	db.On("UpdateDocument", mock.Anything, int64(7), profile).Return(nil)

	c, rec := newEmployeeTestContext(db, httptest.NewRequest(http.MethodDelete, "/api/employees/7", nil))
	c.Params = gin.Params{{Key: "employeeId", Value: "7"}}

	api := &implEmployeeProfilesAPI{}
	api.ArchiveEmployeeProfile(c)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, ARCHIVED, profile.Status)
	require.False(t, profile.ArchivedAt.IsZero())
	require.False(t, profile.UpdatedAt.IsZero())
	db.AssertExpectations(t)
}

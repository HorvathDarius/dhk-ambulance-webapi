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
	"github.com/stretchr/testify/suite"
)

// dbServiceMock implements db_service.DbService[PerformanceRecord] backed by testify mock.
type dbServiceMock struct {
	mock.Mock
}

func (m *dbServiceMock) CreateDocument(ctx context.Context, id any, document *PerformanceRecord) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *dbServiceMock) FindDocument(ctx context.Context, id any) (*PerformanceRecord, error) {
	args := m.Called(ctx, id)
	doc, _ := args.Get(0).(*PerformanceRecord)
	return doc, args.Error(1)
}

func (m *dbServiceMock) UpdateDocument(ctx context.Context, id any, document *PerformanceRecord) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *dbServiceMock) DeleteDocument(ctx context.Context, id any) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *dbServiceMock) ListDocuments(ctx context.Context) ([]*PerformanceRecord, error) {
	args := m.Called(ctx)
	docs, _ := args.Get(0).([]*PerformanceRecord)
	return docs, args.Error(1)
}

func (m *dbServiceMock) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type AmbulanceWlSuite struct {
	suite.Suite
}

func TestAmbulanceWlSuite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	suite.Run(t, new(AmbulanceWlSuite))
}

// newTestContext returns a gin.Context and a recorder with the mock db_service injected.
func (suite *AmbulanceWlSuite) newTestContext(db *dbServiceMock, req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("db_service", db_service.DbService[PerformanceRecord](db))
	c.Request = req
	return c, rec
}

func (suite *AmbulanceWlSuite) Test_GetPerformanceRecord_ReturnsRecord() {
	// ARRANGE
	db := &dbServiceMock{}
	want := &PerformanceRecord{Id: 42, EmployeeName: "MUDr. Test", Date: "2026-04-28", HoursWorked: 8}
	db.On("FindDocument", mock.Anything, int64(42)).Return(want, nil)

	c, rec := suite.newTestContext(db, httptest.NewRequest(http.MethodGet, "/api/performance-records/42", nil))
	c.Params = gin.Params{{Key: "recordId", Value: "42"}}

	// ACT
	api := &implPerformanceRecordsAPI{}
	api.GetPerformanceRecord(c)

	// ASSERT
	suite.Equal(http.StatusOK, rec.Code)

	var got PerformanceRecord
	suite.NoError(json.Unmarshal(rec.Body.Bytes(), &got))
	suite.Equal(want.Id, got.Id)
	suite.Equal(want.EmployeeName, got.EmployeeName)
	db.AssertExpectations(suite.T())
}

func (suite *AmbulanceWlSuite) Test_CreatePerformanceRecord_Returns201WithAssignedId() {
	// ARRANGE
	db := &dbServiceMock{}
	db.On("CreateDocument", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("*ambulance.PerformanceRecord")).Return(nil)

	body := `{"employeeName":"MUDr. Jana","date":"2026-04-28","hoursWorked":8,"examinationCount":1,"operationCount":0,"shiftCount":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/performance-records", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, rec := suite.newTestContext(db, req)

	// ACT
	api := &implPerformanceRecordsAPI{}
	api.CreatePerformanceRecord(c)

	// ASSERT
	suite.Equal(http.StatusCreated, rec.Code)

	var created PerformanceRecord
	suite.NoError(json.Unmarshal(rec.Body.Bytes(), &created))
	suite.NotZero(created.Id, "server should assign a non-zero id")
	suite.Equal("MUDr. Jana", created.EmployeeName)
	db.AssertExpectations(suite.T())
}

func (suite *AmbulanceWlSuite) Test_DeletePerformanceRecord_Returns204OnSuccess() {
	// ARRANGE
	db := &dbServiceMock{}
	db.On("DeleteDocument", mock.Anything, int64(7)).Return(nil)

	c, rec := suite.newTestContext(db, httptest.NewRequest(http.MethodDelete, "/api/performance-records/7", nil))
	c.Params = gin.Params{{Key: "recordId", Value: "7"}}

	// ACT
	api := &implPerformanceRecordsAPI{}
	api.DeletePerformanceRecord(c)

	// ASSERT
	suite.Equal(http.StatusNoContent, rec.Code)
	suite.Empty(rec.Body.Bytes())
	db.AssertExpectations(suite.T())
}

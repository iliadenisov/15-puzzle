package handler_test

import (
	"15-puzzle/internal/model"
	"15-puzzle/internal/repo"
	"15-puzzle/internal/web-service/handler"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	botToken = "example:token"
	initData = "auth_date=269666017&chat_instance=6039284203686499081&chat_type=sender&hash=40cde8dc7250ee616cd7d7a090749a9a42cc68f018c969fcb48fdf5e62657ad6&signature=FF5oTJSnmxdqgtNozxsLywXyVKdssh_DbvksGUaQuhkMiRfp10HJmf5o88uokPpqF4yhpHbX1c8uLbrKUuUdAA&user=%7B%22allows_write_to_pm%22%3Atrue%2C%22first_name%22%3A%22Ilia%22%2C%22id%22%3A303133707%2C%22is_premium%22%3Atrue%2C%22language_code%22%3A%22en%22%2C%22last_name%22%3A%22Denisov%22%2C%22photo_url%22%3A%22https%3A%2F%2Fyoutu.be%2FdQw4w9WgXcQ%22%7D"
	userId   = 303133707
)

func TestHandler(t *testing.T) {
	// static
	testCase(t, testStaticExistingFile)
	testCase(t, testStaticWebAppIndex)
	// api
	testCase(t, testApiInfo)
	testCase(t, testApiAuth)
	testCase(t, testApiStart)
	testCase(t, testApiSolve)
	testCase(t, testApiStats)
	testCase(t, testApiMonitoring)
}

func testStaticExistingFile(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/static/tgwebapp.html", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "static file request must succeed")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "unit-test-header")
}

func testStaticWebAppIndex(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/puzzle.html", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "redirect to static file must succeed")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "unit-test-header")
}

func testApiAuth(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/api/info", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, ctxRoot+"/api/start", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, ctxRoot+"/api/solve?moves=69", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, ctxRoot+"/api/stats", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, ctxRoot+"/api/monitoring", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func testApiInfo(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/api/info", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "api request with correct auth header should be ok")

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var u model.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatalf("decode json %s: %s", w.Body.String(), err)
	}
	assert.NotNil(t, u.Info, "response: stats field should be set")
}

func testApiStart(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, ctxRoot+"/api/start", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "api request with correct auth header should be ok")

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var u model.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatalf("decode json %s: %s", w.Body.String(), err)
	}
	assert.NotNil(t, u.Stats, "response: stats field should be set")
	assert.Equal(t, 1, u.Stats.GamesStarted, "user started games should be exactly one")
}

func testApiSolve(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, ctxRoot+"/api/solve", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, ctxRoot+"/api/solve?moves=69", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "api request with correct auth header should be ok")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var u model.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatalf("decode json %s: %s", w.Body.String(), err)
	}
	assert.NotNil(t, u.Stats, "response: stats field should be set")
	assert.Equal(t, 1, u.Stats.GamesSolved, "user solved games should be exactly one")
}

func testApiStats(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/api/stats", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "api request with correct auth header should be ok")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var u model.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		fmt.Println(w.Body)
		t.Fatalf("decode json %s: %s", w.Body.String(), err)
	}
	assert.NotNil(t, u.Stats, "response: stats field should be set")
}

func testApiMonitoring(t *testing.T, ctxRoot string, h http.Handler) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ctxRoot+"/api/monitoring", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, w.Body.Len() == 0)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, ctxRoot+"/api/monitoring", nil)
	req.Header.Add(handler.WebAppInitDataHeader, initData)
	req.Header.Add(handler.WebAppExtraCodeHeader, "1234")
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var u model.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatalf("decode json %s: %s", w.Body.String(), err)
	}
	assert.NotNil(t, u.Monitoring, "monitoring field should be an object")
	assert.NotNil(t, u.Monitoring.Users)
	assert.NotNil(t, u.Monitoring.GamesStarted)
	assert.NotNil(t, u.Monitoring.GamesSolved)
}

func testCase(t *testing.T, tc func(*testing.T, string, http.Handler)) {
	testContextRoot(t, "", tc)
	testContextRoot(t, "/", tc)
	testContextRoot(t, "/15-puzzle/", tc)
	testContextRoot(t, "/15-puzzle", tc)
}

func testContextRoot(t *testing.T, ctxRoot string, tc func(*testing.T, string, http.Handler)) {
	f, err := os.CreateTemp("", "puzzle15-handler-test")
	if err != nil {
		t.Fatalf("temporary file create: %s", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("temporary file close: %s", err)
	}
	if err := os.Remove(f.Name()); err != nil {
		t.Fatalf("temporary file remove before: %s", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("temporary file remove after: %s", err)
		}
	}()

	r, err := repo.NewFileRepo(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("NewFileRepo: %s", err)
	}
	err = r.AddUser(userId)
	if err != nil {
		t.Fatalf("RegisterGameStart: %s", err)
	}
	tc(t, strings.TrimRight(ctxRoot, "/"), handler.NewHandler(r, botToken, "1234", ctxRoot, "testdata", "projectLink"))
}

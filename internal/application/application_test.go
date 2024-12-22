package application

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Test struct {
	id               int
	method           string
	request          string
	expectedCode     int
	expectedResponse string
}

var tests = []Test{
	{0, http.MethodPost, "{\"expression\": \"2 + 2 * 2\"}", 200, "{\"result\":6}"},
	{1, http.MethodPost, "{\"expression\": \"4 * (3-5 )\"}", 200, "{\"result\":-8}"},
	{2, http.MethodPost, "{\"expression\": \"2 + 2 * 2t\"}", 422, "{\"error\":\"Expression is not valid\"}"},
	{3, http.MethodPost, "{\"expression\": \"2 + r - 2 * 2\"}", 422, "{\"error\":\"Expression is not valid\"}"},
	{4, http.MethodPost, "{\"expression\": \"2 - (2 * 2\"}", 422, "{\"error\":\"Expression is not valid\"}"},
	{5, http.MethodPost, "{\"expression\": \"\"}", 422, "{\"error\":\"Expression is not valid\"}"},
	{6, http.MethodGet, "{\"expression\": \"\"}", 500, "{\"error\":\"Internal server error\"}"},
	{7, http.MethodPost, "{\"exression\": \"2 + 2\"}", 422, "{\"error\":\"Expression is not valid\"}"},
	{8, http.MethodPost, "\"expression\": \"2 + 2\"}", 500, "{\"error\":\"Internal server error\"}"},
	{9, http.MethodPost, "{\"expression\": \"2 + 2\"}, \"s\": \"f\"", 200, "{\"result\":4}"},
}

func TestCalcHandler(t *testing.T) {
	for _, test := range tests {
		_TestCalcHandler(t, &test)
	}
}

func _TestCalcHandler(t *testing.T, test *Test) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic in test #%d - %s", test.id, r)
		}
	}()

	w := httptest.NewRecorder()

	r := httptest.NewRequest(test.method, "/api/v1/calculate", strings.NewReader(test.request))
	r.Header.Set("Content-Type", "application/json")

	CalcHandler(w, r)

	result := w.Result()
	if result.StatusCode != test.expectedCode {
		t.Errorf("#%d: expected status %d, got %d", test.id, test.expectedCode, result.StatusCode)
	}

	response, _ := io.ReadAll(result.Body)
	if strings.TrimSpace(string(response)) != test.expectedResponse {
		t.Errorf("#%d: expected response \"%s\", got \"%s\"", test.id, test.expectedResponse, string(response))
	}
}

package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/setup"
	"github.com/gofiber/fiber/v2"
)

// Prerequisites:
// running PostgreSQL and Redis instance
// test database as in setup/setupTestDatabase
// users: Admin(id=1,password=password1234), User(id=2,password=password1234)
// roles: admin(id=1), test=(id=2)
// user_roles: user Admin(id=1) has role admin(id=1)
// user_registration-keys: user User(id=2) is registered with Registration-Key test_registration_key(id=1)
// registration_keys: test_registration_key(id=1,permanent)
// TESTUSER=Admin,TESTPASSWORD=password1234

const (
	URL          = "http://localhost:8080"
	TESTUSER     = "Admin"
	TESTPASSWORD = "password1234"
)

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
}

func expect2xxStatus(t *testing.T, resp *http.Response) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Bad status code: Expected 2xx, got %d", resp.StatusCode)
	}
}

func expect404Status(t *testing.T, resp *http.Response) {
	if resp.StatusCode != 404 {
		t.Fatalf("Bad status code: Expected %d, got %d", 404, resp.StatusCode)
	}
}

func readBodyAsJson(t *testing.T, resp *http.Response, jsonType any) {
	body, err := io.ReadAll(resp.Body)
	checkError(t, err)
	err = json.Unmarshal(body, jsonType)
	checkError(t, err)
}

func payloadToReader(t *testing.T, payload any) *bytes.Reader {
	bs, err := json.Marshal(payload)
	checkError(t, err)
	return bytes.NewReader(bs)
}

func login(t *testing.T, app *fiber.App) string {
	payload := handler.LoginPayload{
		Username: TESTUSER,
		Password: TESTPASSWORD,
	}
	req, err := http.NewRequest("POST", URL+"/login", payloadToReader(t, payload))
	checkError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp, err := app.Test(req)
	checkError(t, err)
	expect2xxStatus(t, resp)
	cookie := resp.Header.Get("Set-Cookie")
	cookie = strings.Split(cookie, ";")[0]
	return cookie
}

func RunRequest(t *testing.T, req *http.Request) *http.Response {
	// start := time.Now()
	app := setup.Setup()
	// t.Logf("Setup time: %v", time.Since(start))
	cookie := login(t, app)

	req.Header.Add("Cookie", cookie)
	if req.Body != http.NoBody {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	checkError(t, err)
	return resp
}

func RunMultiRequest(t *testing.T, reqs ...*http.Request) []*http.Response {
	app := setup.Setup()
	cookie := login(t, app)
	resps := make([]*http.Response, len(reqs))
	for i, req := range reqs {
		req.Header.Add("Cookie", cookie)
		if req.Body != http.NoBody {
			req.Header.Add("Content-Type", "application/json")
		}
		resp, err := app.Test(req)
		checkError(t, err)
		resps[i] = resp
	}
	return resps
}

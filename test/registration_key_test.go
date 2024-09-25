package test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/model"
)

func TestGetRegistrationKeys(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/registration-keys", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var regKeys []model.RegistrationKey
	readBodyAsJson(t, resp, &regKeys)
	t.Logf("Got registration-keys: %+v", regKeys)
}

func TestGetRegistrationKeyByKey(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/registration-keys?key=test_registration_key", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var regKey model.RegistrationKey
	readBodyAsJson(t, resp, &regKey)
	t.Logf("Got registration-key: %+v", regKey)
}

func TestGetRegistrationKeyByKeyThatDoesNotExist(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/registration-keys?key=doesnotexist", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect404Status(t, resp)
}

func TestCreateRegistrationKey(t *testing.T) {
	payload := handler.CreateRegistrationKeyPayload{
		Key:         "NewRegKey123",
		Description: "TestKey",
		Permanent:   false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	req1, err := http.NewRequest("POST", URL+"/registration-keys", payloadToReader(t, payload))
	checkError(t, err)

	// check if key exists
	req2, err := http.NewRequest("GET", URL+"/registration-keys?key=NewRegKey123", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])
}

func TestUpdateRegistrationKey(t *testing.T) {
	payload := handler.UpdateRegistrationKeyPayload{
		Description: "TestNewDescription",
		Permanent:   false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	req1, err := http.NewRequest("PUT", URL+"/registration-keys/1", payloadToReader(t, payload))
	checkError(t, err)

	req2, err := http.NewRequest("GET", URL+"/registration-keys/1", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])

	var regKey model.RegistrationKey
	readBodyAsJson(t, resps[1], &regKey)
	if regKey.Description != "TestNewDescription" {
		t.Fatal("Registration key description was not updated")
	}
	if regKey.Permanent {
		t.Fatal("Registration key permanent was not updated")
	}
}

// TODO: delete doesn't work when user is associated (foreign key constraint violation)
func TestDeleteRegistrationKey(t *testing.T) {
	req1, err := http.NewRequest("DELETE", URL+"/registration-keys/1", nil)
	checkError(t, err)
	req2, err := http.NewRequest("GET", URL+"/registration-keys/1", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect404Status(t, resps[1])
}

func TestGetUsersOfRegistrationKey(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/registration-keys/1/users", nil)
	checkError(t, err)
	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)
	var users []model.User
	readBodyAsJson(t, resp, &users)
	t.Logf("Got users: %v", users)
}

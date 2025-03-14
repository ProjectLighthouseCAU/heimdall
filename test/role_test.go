package test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/model"
)

func TestGetRoles(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var roles []model.Role
	readBodyAsJson(t, resp, &roles)
	t.Logf("Got roles: %+v", roles)
}

func TestGetRoleByName(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles?name=admin", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var role model.Role
	readBodyAsJson(t, resp, &role)
	t.Logf("Got role: %+v", role)
}

func TestGetRoleByNameThatDoesNotExist(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles?name=doesnotexist", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect404Status(t, resp)
}

func TestGetRoleById(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles/1", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var role model.Role
	readBodyAsJson(t, resp, &role)
	t.Logf("Got role: %v", role)
}

func TestGetRoleByIdThatDoesNotExist(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles/3", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect404Status(t, resp)
}

func TestCreateRole(t *testing.T) {
	payload := handler.CreateOrUpdateRolePayload{
		Name: "Testrole",
	}
	req1, err := http.NewRequest("POST", URL+"/roles", payloadToReader(t, payload))
	checkError(t, err)

	// check if role exists
	req2, err := http.NewRequest("GET", URL+"/roles?name=Testrole", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])
}

func TestUpdateRole(t *testing.T) {
	payload := handler.CreateOrUpdateRolePayload{
		Name: "test_updated",
	}
	req1, err := http.NewRequest("PUT", URL+"/roles/2", payloadToReader(t, payload))
	checkError(t, err)

	// check if role exists
	req2, err := http.NewRequest("GET", URL+"/roles/2", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])
	var role model.Role
	readBodyAsJson(t, resps[1], &role)
	if role.Name != "test_updated" {
		t.Fatalf("Role name was not correctly updated: Expected name: test_updated, got name: %s", role.Name)
	}
}

func TestGetUsersOfRole(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/roles/1/users", nil)
	checkError(t, err)

	resp := RunRequest(t, req)
	expect2xxStatus(t, resp)

	var users []model.User
	readBodyAsJson(t, resp, &users)
	t.Logf("Got users: %+v", users)
}

func TestAddUserToRole(t *testing.T) {
	req1, err := http.NewRequest("PUT", URL+"/roles/2/users/3", nil)
	checkError(t, err)

	req2, err := http.NewRequest("GET", URL+"/roles/2/users", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])

	var users []model.User
	readBodyAsJson(t, resps[1], &users)
	if !slices.ContainsFunc(users, func(user model.User) bool {
		return user.Username == "User"
	}) {
		t.Fatalf("Role does not contain added user \"User\", only %+v", users)
	}
}

func TestRemoveUserFromRole(t *testing.T) {
	req1, err := http.NewRequest("PUT", URL+"/roles/2/users/3", nil)
	checkError(t, err)

	req2, err := http.NewRequest("DELETE", URL+"/roles/2/users/3", nil)
	checkError(t, err)

	req3, err := http.NewRequest("GET", URL+"/roles/2/users", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2, req3)
	expect2xxStatus(t, resps[0])
	expect2xxStatus(t, resps[1])
	expect2xxStatus(t, resps[2])

	var users []model.User
	readBodyAsJson(t, resps[2], &users)
	if slices.ContainsFunc(users, func(user model.User) bool {
		return user.Username == "User"
	}) {
		t.Fatalf("Role contains added user \"User\" after removal, users: %+v", users)
	}
}

func TestDeleteRole(t *testing.T) {
	req1, err := http.NewRequest("DELETE", URL+"/roles/2", nil)
	checkError(t, err)

	// check if role exists
	req2, err := http.NewRequest("GET", URL+"/roles/2", nil)
	checkError(t, err)

	resps := RunMultiRequest(t, req1, req2)
	expect2xxStatus(t, resps[0])
	expect404Status(t, resps[1])
}

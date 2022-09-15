package auth

import (
	"testing"

	"github.com/casbin/casbin/v2"
)

// TODO: make sure that users can't name themselves role::something or :username
// TODO: use SyncedEnforcer or CachedEnforcer in non-test code
func TestCasbin(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("../casbin/model.conf", "../casbin/policy.csv")
	if err != nil {
		t.Fatalf("Error creating enforcer: %v", err)
	}
	type testcase struct {
		subject string
		object  string
		action  string
		result  bool
	}
	tests := []testcase{
		// users should not read or write other users resources
		{"testuser", "/user/testadmin/model", "read", false},
		{"testuser", "/user/testadmin/model", "write", false},
		// users should read and write own resources
		{"testuser", "/user/testuser/model", "read", true},
		{"testuser", "/user/testuser/model", "write", true},
		// users should read but not write other users resources if set to public
		{"testuser", "/user/testpublicuser/model", "read", true},
		{"testuser", "/user/testpublicuser/model", "write", false},
		// admins should read and write other users resources
		{"testadmin", "/user/testuser/model", "read", true},
		{"testadmin", "/user/testuser/model", "write", true},
	}
	for _, test := range tests {
		ok, err := enforcer.Enforce(test.subject, test.object, test.action)
		if err != nil {
			t.Fatalf("Error enforcing: %v", err)
		}
		if ok != test.result {
			t.Errorf("{%v %v %v}: Expected %v, got %v", test.subject, test.object, test.action, test.result, ok)
		}
	}
	enforcer.AddRoleForUser("otheruser", "testrole")

	enforcer.SavePolicy()
}

package access

import (
	"fmt"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/neuvector/neuvector/controller/api"
	"github.com/neuvector/neuvector/share"
	"github.com/neuvector/neuvector/share/utils"
)

func preTest() {
	log.SetLevel(log.FatalLevel)

	CompileUriPermitsMapping()
}

func postTest() {
	log.SetLevel(log.DebugLevel)
}

type globalObject struct{}

func (o *globalObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return nil, nil
}

type domainObject struct{}

func (o *domainObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return []string{"ns1", "nsx"}, nil
}

type domainObjectTest struct {
	CreatorDomains []string
}

func (o *domainObjectTest) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return o.CreatorDomains, nil
}

type dualObject struct{}

func (o *dualObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return []string{"ns1", "nsx"}, []string{"ns2", "nsy"}
}

type allReaderObject struct{}

func (o *allReaderObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return []string{share.AccessAllAsReader}, nil
}

func TestGlobalAccess(t *testing.T) {
	preTest()

	var obj globalObject
	r, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/controller/12345", nil)

	acc := NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleAdmin,
	})
	authz := acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	postTest()
}

func TestDomainAccess(t *testing.T) {
	preTest()

	var obj domainObject
	r, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/workload/12345", nil)

	acc := NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleAdmin,
	})
	authz := acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"":    api.UserRoleReader,
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"":    api.UserRoleReader,
		"ns2": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns2": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	postTest()
}

func TestWildcardDomainAccess(t *testing.T) {
	preTest()

	obj1 := domainObjectTest{
		CreatorDomains: []string{"ns-dev-*", "ns1"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/workload/12345", nil)
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		domains := []string{"", "ns-dev-*", "ns-dev-**", "ns1"}
		for _, domain := range domains {
			acc := NewAccessControl(req, op, DomainRole{
				domain: userRole1.Name,
			})
			authz := acc.Authorize(&obj1, nil)
			if op == AccessOPRead {
				if !authz {
					t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
				}
			} else {
				if authz {
					t.Errorf("Authz success: op=%v, roles=%+v", acc.op, acc.roles)
				}
			}
		}
		for _, domain := range domains {
			acc := NewAccessControl(req, op, DomainRole{
				domain:      userRole1.Name,
				"default-2": userRole1.Name,
			})
			authz := acc.Authorize(&obj1, nil)
			if op == AccessOPRead {
				if !authz {
					t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
				}
			} else {
				if authz {
					t.Errorf("Authz success: op=%v, roles=%+v", acc.op, acc.roles)
				}
			}
		}

		domains = []string{"ns-dev", "ns-dev-A", "ns-dev-", "ns-dev-*-*", "ns-dev-b1-*", "ns", "ns2", "ns11"}
		for _, domain := range domains {
			acc := NewAccessControl(req, op, DomainRole{
				domain: userRole1.Name,
			})
			authz := acc.Authorize(&obj1, nil)
			if authz {
				t.Errorf("Authz succeed: op=%v, roles=%+v", acc.op, acc.roles)
			}
		}
		for _, domain := range domains {
			acc := NewAccessControl(req, op, DomainRole{
				domain:      userRole1.Name,
				"default-2": userRole1.Name,
			})
			authz := acc.Authorize(&obj1, nil)
			if authz {
				t.Errorf("Authz succeed: op=%v, roles=%+v", acc.op, acc.roles)
			}
		}
		for _, domain := range domains {
			acc := NewAccessControl(req, op, DomainRole{
				domain: userRole1.Name,
				"":     userRole1.Name,
			})
			authz := acc.Authorize(&obj1, nil)
			if op == AccessOPRead {
				if !authz {
					t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
				}
			} else {
				if authz {
					t.Errorf("Authz succeed: op=%v, roles=%+v", acc.op, acc.roles)
				}
			}
		}
	}

	postTest()
}

type UserDomainsResult struct {
	UserDomains    []string
	ExpectedResult bool
}

func testWildcardDomainAccess(caller string, t *testing.T, obj *domainObjectTest, req *http.Request, op AccessOP,
	userRole1, userRole2 *share.CLUSUserRoleInternal, userDomainsResultList []UserDomainsResult) {
	// role-1 user testing
	for idx, userDomainsResult := range userDomainsResultList {
		domainRole := map[string]string{}
		for _, domain := range userDomainsResult.UserDomains {
			domainRole[domain] = userRole1.Name
		}
		acc := NewAccessControl(req, op, domainRole)
		authz := acc.Authorize(obj, nil)
		if op == AccessOPRead {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for read fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for read success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		} else {
			if authz {
				t.Errorf("[%s:%d] Authz for write success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		}
	}
	// role-2 user testing
	for idx, userDomainsResult := range userDomainsResultList {
		domainRole := map[string]string{}
		for _, domain := range userDomainsResult.UserDomains {
			domainRole[domain] = userRole2.Name
		}
		acc := NewAccessControl(req, op, domainRole)
		authz := acc.Authorize(obj, nil)
		if op == AccessOPRead {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for read fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for read success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		} else {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for write fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for write success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		}
	}
}

// Example 1: user A1 has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*" only and creates a group G1 ( i.e. group G1's CreaterDomains is ["app-US-*"]	)
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1"          cannot access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1", "app-2" cannot access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-JP-*"       cannot access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*"       can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*", "qa" can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-*", "qa"    can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-*"          can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "", "qa"         can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain ""               can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "*"              can access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-dev-*"   cannot access group G1
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "*-US-dev" ,"*b" cannot access group G1
func TestWildcardDomainAccess1(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-US-*"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1", "app-2"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-*", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-*", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-dev-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"*-US-dev", "*b"}, ExpectedResult: false},
		}
		testWildcardDomainAccess("TestWildcardDomainAccess1", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

// Example 2: user A2 has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1", "app-US-dev" and creates a group G2 ( i.e. group G2's CreaterDomains is ["app-1", "app-US-dev"]	)
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1"          		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-dev"     		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1", "qa"    		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1", "app-US-dev" 	can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-*"          		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*"       		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "","app-JP"       		can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain ""       				can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "*"       				can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "*-dev"       			can access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US"       			cannot access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-2","app-JP-*"       cannot access group G2
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "*-dev-US"       		cannot access group G2
func TestWildcardDomainAccess2(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-1", "app-US-dev"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-dev"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "app-US-dev"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"", "app-JP"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*-dev"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-2", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"*-dev-US"}, ExpectedResult: false},
		}
		testWildcardDomainAccess("TestWildcardDomainAccess2", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

// Example 3: user A3 has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1", "app-US-*" and creates a group G3 ( i.e. group G3's CreaterDomains is ["app-1", "app-US-*"]	)
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1"          		can access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*"       		can access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app*"          			can access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-1","qa"     		can access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-*","app-JP-*" 	can access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-1" 				cannot access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-US-1","app-JP-2" 	cannot access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "bpp-*"          		cannot access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "qa" 					cannot access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "qa","app-JP-*" 			cannot access group G3
// 	user has PERMS_RUNTIME_POLICIES(r/w) on domain "app-UK-1","app-JP-2" 	cannot access group G3
func TestWildcardDomainAccess3(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-1", "app-US-*"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-*", "app-JP-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-US-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-1", "app-JP-2"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"bpp-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"qa"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"qa", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-UK-1", "app-JP-2"}, ExpectedResult: false},
		}
		testWildcardDomainAccess("TestWildcardDomainAccess3", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

func TestWildcardDomainAccess4(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: nil,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1", "app-2"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*", "qa"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-*", "qa"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"", "qa"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: false}, // { role -> "*" } still cannot access global namespace !
			UserDomainsResult{UserDomains: []string{"app-US-dev-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"*-US-dev", "*b"}, ExpectedResult: false},
		}
		testWildcardDomainAccess("TestWildcardDomainAccess4", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

func TestDualAccess(t *testing.T) {
	preTest()

	var obj dualObject
	r, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/scan/workload/12345", nil)

	acc := NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleAdmin,
	})
	authz := acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
		"ns2": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
		"ns2": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleAdmin,
		"ns2": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleReader,
		"ns2": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	// For read access, readable to one side of domain list is enough
	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns2": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	postTest()
}

// --

type readObject struct {
	members []string
}

func newReadObject(members []string) *readObject {
	return &readObject{members: members}
}

func (o *readObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return o.members, nil
}

// --

type ownObject struct {
	members []string
}

func newOwnObject(members []string) *ownObject {
	return &ownObject{members: members}
}

func (o *ownObject) GetDomain(f share.GetAccessObjectFunc) ([]string, []string) {
	return o.members, nil
}

func TestOwnAccess(t *testing.T) {
	preTest()

	obj := newOwnObject(nil)
	r, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/user/12345", nil)

	acc := NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleAdmin,
	})
	authz := acc.AuthorizeOwn(obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	obj = newOwnObject([]string{"ns1", "ns2"})

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleReader,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
		"ns2": api.UserRoleReader,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
		"ns2": api.UserRoleAdmin,
	})
	authz = acc.AuthorizeOwn(obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	postTest()
}

func TestWildcardOwnAccess(t *testing.T) {
	preTest()

	obj1 := ownObject{
		members: []string{"ns-dev-*", "ns1"},
	}
	req, _ := http.NewRequest(http.MethodPatch, "https://10.1.1.1/v1/workload/12345", nil)
	userRole1 := &share.CLUSUserRoleInternal{
		Name:         "role-1",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		domainss := [][]string{[]string{""}, []string{"ns*"}, []string{"ns-*", "ns1"}, []string{"ns1", "ns-qa-*", "ns-dev-*"}}
		for _, domains := range domainss {
			domainRole := DomainRole{}
			for _, domain := range domains {
				domainRole[domain] = userRole1.Name
			}
			acc := NewAccessControl(req, op, domainRole)
			authz := acc.AuthorizeOwn(&obj1, nil)
			if !authz {
				t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
			}
		}
	}

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		domainss := [][]string{[]string{"ns-dev-*"}, []string{"ns1*"}, []string{"ns-dev-*", "ns2"}, []string{"ns-dev1-*", "ns1"}}
		for _, domains := range domainss {
			domainRole := DomainRole{}
			for _, domain := range domains {
				domainRole[domain] = userRole1.Name
			}
			acc := NewAccessControl(req, op, domainRole)
			authz := acc.AuthorizeOwn(&obj1, nil)
			if authz {
				t.Errorf("Authz success: op=%v, roles=%+v", acc.op, acc.roles)
			}
		}
	}

	postTest()
}

func testWildcardOwnAccess(caller string, t *testing.T, obj *domainObjectTest, req *http.Request, op AccessOP,
	userRole1, userRole2 *share.CLUSUserRoleInternal, userDomainsResultList []UserDomainsResult) {
	// role-1 user testing
	for idx, userDomainsResult := range userDomainsResultList {
		domainRole := map[string]string{}
		for _, domain := range userDomainsResult.UserDomains {
			domainRole[domain] = userRole1.Name
		}
		acc := NewAccessControl(req, op, domainRole)
		// Authorize if the access has rights on all domains which obj1 is member of.
		authz := acc.AuthorizeOwn(obj, nil)
		if op == AccessOPRead {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for read fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for read success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		} else {
			if authz {
				t.Errorf("[%s:%d] Authz for write success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		}
	}
	// role-2 user testing
	for idx, userDomainsResult := range userDomainsResultList {
		domainRole := map[string]string{}
		for _, domain := range userDomainsResult.UserDomains {
			domainRole[domain] = userRole2.Name
		}
		acc := NewAccessControl(req, op, domainRole)
		// Authorize if the access has rights on all domains which obj1 is member of.
		authz := acc.AuthorizeOwn(obj, nil)
		if op == AccessOPRead {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for read fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for read success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		} else {
			if userDomainsResult.ExpectedResult && !authz {
				t.Errorf("[%s:%d] Authz for write fail: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			} else if !userDomainsResult.ExpectedResult && authz {
				t.Errorf("[%s:%d] Authz for write success: op=%v, roles=%+v", caller, idx, acc.op, acc.roles)
			}
		}
	}
}

func TestWildcardOwnAccess1(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-1", "app-US-*"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1*", "app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "app-US-*", "app-JP-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-2", "app-US-*", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: false},
		}
		testWildcardOwnAccess("TestWildcardOwnAccess1", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

func TestWildcardOwnAccess2(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-1", "app-US-1"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1*", "app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "app-US-1", "app-JP-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-2", "app-US-*", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: false},
		}
		testWildcardOwnAccess("TestWildcardOwnAccess2", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

func TestWildcardOwnAccess3(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: []string{"app-US-*"},
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-JP*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1*", "app-US-*"}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"app-1", "app-US-CA-*", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-2", "app-US-1", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*-temp"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: true},
		}
		testWildcardOwnAccess("TestWildcardOwnAccess3", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

func TestWildcardOwnAccess4(t *testing.T) {
	preTest()

	obj := domainObjectTest{
		CreatorDomains: nil,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/group/12345", nil) // method can be ignored because we use different op for testing later
	userRole1 := &share.CLUSUserRoleInternal{
		Name:        "role-1",
		ReadPermits: share.PERMS_RUNTIME_POLICIES,
	}
	userRole2 := &share.CLUSUserRoleInternal{
		Name:         "role-2",
		ReadPermits:  share.PERMS_RUNTIME_POLICIES,
		WritePermits: share.PERMS_RUNTIME_POLICIES,
	}
	AddRole(userRole1.Name, userRole1)
	AddRole(userRole2.Name, userRole2)

	for _, op := range []AccessOP{AccessOPWrite, AccessOPRead} {
		userDomainsResultList := []UserDomainsResult{
			UserDomainsResult{UserDomains: []string{""}, ExpectedResult: true},
			UserDomainsResult{UserDomains: []string{"*"}, ExpectedResult: false}, // { role -> "*" } still cannot access global namespace !
			UserDomainsResult{UserDomains: []string{"app*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1*", "app-US-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1", "app-US-1", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-2", "app-US-*", "app-JP-*"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-1"}, ExpectedResult: false},
			UserDomainsResult{UserDomains: []string{"app-US-*"}, ExpectedResult: false},
		}
		testWildcardOwnAccess("TestWildcardOwnAccess4", t, &obj, req, op, userRole1, userRole2, userDomainsResultList)
	}

	postTest()
}

// --

func TestAllReaderAccess(t *testing.T) {
	preTest()

	var obj allReaderObject
	r, _ := http.NewRequest(http.MethodGet, "https://10.1.1.1/v1/scan/registry/12345", nil)

	acc := NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleAdmin,
	})
	authz := acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"": api.UserRoleCIOps,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPWrite, DomainRole{
		"ns1": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleAdmin,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleReader,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != true {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{
		"ns1": api.UserRoleCIOps,
	})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	acc = NewAccessControl(r, AccessOPRead, DomainRole{})
	authz = acc.Authorize(&obj, nil)
	if authz != false {
		t.Errorf("Authz fail: op=%v, roles=%+v", acc.op, acc.roles)
	}

	postTest()
}

func checkPermissionValue(id string, pValue uint64, t *testing.T) {
	original := pValue
	var bit uint64 = 0x8000000000000000
	found := false
	for ; bit > 0; bit >>= 1 {
		if bit&pValue > 0 {
			if found {
				t.Errorf("Found simple permission with non-simple value: id=%v, value=0x%x", id, original)
				return
			}
			found = true
		}
		pValue = pValue & (bit - 1)
	}
}

func TestSimplePermission(t *testing.T) {
	preTest()

	for _, option := range PermissionOptions {
		if option.ComplexPermits == nil {
			checkPermissionValue(option.ID, option.Value, t)
		} else {
			for _, option2 := range option.ComplexPermits {
				checkPermissionValue(option.ID, option2.Value, t)
			}
		}
	}

	readOnlyPermissions := utils.NewSet(share.PERM_AUDIT_EVENTS, share.PERM_SECURITY_EVENTS_BASIC, share.PERM_EVENTS)
	writeOnlyPermissions := utils.NewSet(share.PERM_CICD_SCAN)
	globalOnlyPermissions := utils.NewSet(share.PERM_CICD_SCAN, share.PERM_NV_RESOURCE, share.PERM_ADM_CONTROL, share.PERM_AUTHENTICATION, share.PERM_SYSTEM_CONFIG,
		share.PERM_CLOUD, share.PERM_INFRA_BASIC, share.PERM_VULNERABILITY)

	globalReadPermissions := []uint64{share.PERM_NV_RESOURCE, share.PERM_RUNTIME_SCAN_BASIC, share.PERM_REG_SCAN, share.PERM_NETWORK_POLICY_BASIC, share.PERM_SYSTEM_POLICY_BASIC,
		share.PERM_GROUP_BASIC, share.PERM_ADM_CONTROL, share.PERM_COMPLIANCE_BASIC, share.PERM_AUDIT_EVENTS, share.PERM_SECURITY_EVENTS_BASIC, share.PERM_EVENTS, share.PERM_AUTHENTICATION,
		share.PERM_AUTHORIZATION, share.PERM_SYSTEM_CONFIG, share.PERM_CLOUD, share.PERM_WORKLOAD_BASIC, share.PERM_INFRA_BASIC, share.PERM_VULNERABILITY}

	globalWritePermissions := []uint64{share.PERM_NV_RESOURCE, share.PERM_RUNTIME_SCAN_BASIC, share.PERM_REG_SCAN, share.PERM_NETWORK_POLICY_BASIC, share.PERM_SYSTEM_POLICY_BASIC,
		share.PERM_GROUP_BASIC, share.PERM_ADM_CONTROL, share.PERM_COMPLIANCE_BASIC, share.PERM_AUTHENTICATION, share.PERM_AUTHORIZATION, share.PERM_SYSTEM_CONFIG,
		share.PERM_CLOUD, share.PERM_WORKLOAD_BASIC, share.PERM_INFRA_BASIC, share.PERM_CICD_SCAN, share.PERM_VULNERABILITY}

	domainReadPermissions := []uint64{share.PERM_RUNTIME_SCAN_BASIC, share.PERM_REG_SCAN, share.PERM_NETWORK_POLICY_BASIC, share.PERM_SYSTEM_POLICY_BASIC, share.PERM_GROUP_BASIC,
		share.PERM_COMPLIANCE_BASIC, share.PERM_AUTHORIZATION, share.PERM_SYSTEM_CONFIG, share.PERM_WORKLOAD_BASIC, share.PERM_AUDIT_EVENTS, share.PERM_SECURITY_EVENTS_BASIC, share.PERM_EVENTS}

	domainWritePermissions := []uint64{share.PERM_RUNTIME_SCAN_BASIC, share.PERM_REG_SCAN, share.PERM_NETWORK_POLICY_BASIC, share.PERM_SYSTEM_POLICY_BASIC,
		share.PERM_GROUP_BASIC, share.PERM_COMPLIANCE_BASIC, share.PERM_AUTHORIZATION, share.PERM_WORKLOAD_BASIC}

	for _, p := range globalReadPermissions {
		if writeOnlyPermissions.Contains(p) {
			t.Errorf("Found read permission in write only permissions=0x%x", p)
		}

		if (p & share.PERMS_CLUSTER_READ) != p {
			t.Errorf("Not found read permission in global read permissions=0x%x", p)
		}
	}

	for _, p := range globalWritePermissions {
		if readOnlyPermissions.Contains(p) {
			t.Errorf("Found write permission in read only permissions=0x%x", p)
		}

		if (p & share.PERMS_CLUSTER_WRITE) != p {
			t.Errorf("Not found write permission in global write permissions=0x%x", p)
		}
	}
	for _, p := range domainReadPermissions {
		if writeOnlyPermissions.Contains(p) {
			t.Errorf("Found read permission in write only permissions=0x%x", p)
		}
		if globalOnlyPermissions.Contains(p) {
			t.Errorf("Found global permission in domain read permissions=0x%x", p)
		}

		if (p & share.PERMS_DOMAIN_READ) != p {
			t.Errorf("Not found read permission in domain read permissions=0x%x", p)
		}
	}
	for _, p := range domainWritePermissions {
		if readOnlyPermissions.Contains(p) {
			t.Errorf("Found write permission in read only permissions=0x%x", p)
		}
		if globalOnlyPermissions.Contains(p) {
			t.Errorf("Found global permission in domain write permissions=0x%x", p)
		}

		if (p & share.PERMS_DOMAIN_WRITE) != p {
			t.Errorf("Not found write  permission in domain write permissions=0x%x", p)
		}
	}

	if (share.PERMS_CLUSTER_READ - share.PERMS_GLOBAL_CONFIGURABLE_READ) != share.PERM_NV_RESOURCE {
		t.Errorf("wrong configurable read permissions=(0x%x : 0x%x)", share.PERMS_GLOBAL_CONFIGURABLE_READ, share.PERMS_CLUSTER_READ)
	}
	if (share.PERMS_CLUSTER_WRITE - share.PERMS_GLOBAL_CONFIGURABLE_WRITE) != share.PERM_NV_RESOURCE {
		t.Errorf("wrong configurable write permissions=(0x%x : 0x%x)", share.PERMS_GLOBAL_CONFIGURABLE_WRITE, share.PERMS_CLUSTER_WRITE)
	}

	postTest()
}

func TestCompileApiUrisMappingMapping(t *testing.T) {
	preTest()

	apiURIsGET := map[int8][]string{
		CONST_API_NO_AUTH: []string{
			"v1/partner/ibm_sa/*/setup",
			"v1/partner/ibm_sa/*/setup/*",
			"v1/token_auth_server",
			"v1/token_auth_server/*",
			"v1/eula",
			"v1/fed/healthcheck",
		},
		CONST_API_DEBUG: []string{
			"v1/fed/member",
			"v1/meter",
			"v1/enforcer/*/probe_summary",
			"v1/enforcer/*/probe_processes",
			"v1/enforcer/*/probe_containers",
			"v1/debug/ip2workload",
			"v1/debug/internal_subnets",
			"v1/debug/policy/rule",
			"v1/debug/dlp/wlrule",
			"v1/debug/dlp/rule",
			"v1/debug/dlp/mac",
			"v1/debug/system/stats",
			"v1/debug/controller/sync",
			"v1/debug/workload/intercept",
			"v1/debug/registry/image/*",
			"v1/session/summary",
			"v1/file_monitor_file",
			"v1/system/usage",
			"v1/system/rbac",
		},
		CONST_API_RT_SCAN: []string{
			"v1/scan/config",
			"v1/scan/status",
			"v1/scan/workload/*",
			"v1/scan/image",
			"v1/scan/image/*",
			"v1/scan/host/*",
			"v1/scan/platform",
			"v1/scan/platform/platform",
			"v1/scan/asset",
			"v1/vulasset",
		},
		CONST_API_REG_SCAN: []string{
			"v1/scan/registry",
			"v1/scan/registry/*",
			"v1/scan/registry/*/images",
			"v1/scan/registry/*/image/*",
			"v1/scan/registry/*/layers/*",
			"v1/list/registry_type",
			"v1/scan/sigstore/root_of_trust",
			"v1/scan/sigstore/root_of_trust/*",
			"v1/scan/sigstore/root_of_trust/*/verifier",
			"v1/scan/sigstore/root_of_trust/*/verifier/*",
		},
		CONST_API_INFRA: []string{
			"v1/host",
			"v1/host/*",
			"v1/host/*/process_profile",
			"v1/domain",
		},
		CONST_API_NV_RESOURCE: []string{
			"v1/controller",
			"v1/controller/*",
			"v1/controller/*/config",
			"v1/controller/*/stats",
			"v1/controller/*/counter",
			"v1/enforcer",
			"v1/enforcer/*",
			"v1/enforcer/*/stats",
			"v1/enforcer/*/counter",
			"v1/enforcer/*/config",
			"v1/scan/scanner",
		},
		CONST_API_WORKLOAD: []string{
			"v1/workload",
			"v2/workload",
			"v1/workload/*",
			"v2/workload/*",
			"v1/workload/*/stats",
			"v1/workload/*/config",
		},
		CONST_API_GROUP: []string{
			"v1/group",
			"v1/group/*",
			"v1/service",
			"v1/service/*",
			"v1/file/group",
		},
		CONST_API_RT_POLICIES: []string{
			"v1/workload/*/process",
			"v1/workload/*/process_history",
			"v1/workload/*/process_profile",
			"v1/workload/*/file_profile",
			"v1/dlp/sensor",
			"v1/dlp/sensor/*",
			"v1/dlp/group",
			"v1/dlp/group/*",
			"v1/dlp/rule",
			"v1/dlp/rule/*",
			"v1/waf/sensor",
			"v1/waf/sensor/*",
			"v1/waf/group",
			"v1/waf/group/*",
			"v1/waf/rule",
			"v1/waf/rule/*",
			"v1/policy/rule",
			"v1/policy/rule/*",
			"v1/session",
			"v1/conversation_endpoint",
			"v1/conversation",
			"v1/conversation/*/*",
			"v1/process_profile",
			"v1/process_profile/*",
			"v1/process_rules/*",
			"v1/file_monitor",
			"v1/file_monitor/*",
			"v1/response/rule",
			"v1/response/rule/*",
			"v1/response/options",
			"v1/response/workload_rules/*",
			"v1/list/application",
			"v1/sniffer",
			"v1/sniffer/*",
			"v1/sniffer/*/pcap",
			"v1/file/group/config",
		},
		CONST_API_ADM_CONTROL: []string{
			"v1/admission/options",
			"v1/admission/state",
			"v1/admission/stats",
			"v1/admission/rules",
			"v1/admission/rule/*",
			"v1/debug/admission_stats",
		},
		CONST_API_COMPLIANCE: []string{
			"v1/host/*/compliance",
			"v1/workload/*/compliance",
			"v1/bench/host/*/docker",
			"v1/bench/host/*/kubernetes",
			"v1/custom_check/*",
			"v1/custom_check",
			"v1/compliance/asset",
			"v1/list/compliance",
			"v1/compliance/profile",
			"v1/compliance/profile/*",
		},
		CONST_API_AUDIT_EVENTS: []string{
			"v1/log/audit",
		},
		CONST_API_SECURITY_EVENTS: []string{
			"v1/log/incident",
			"v1/log/threat",
			"v1/log/threat/*",
			"v1/log/violation",
			"v1/log/security",
			"v1/log/violation/workload",
		},
		CONST_API_EVENTS: []string{
			"v1/log/event",
			"v1/log/activity",
		},
		CONST_API_AUTHENTICATION: []string{
			"v1/server",
			"v1/server/*",
			"v1/server/*/user",
		},
		CONST_API_AUTHORIZATION: []string{
			"v1/user_role_permission/options",
			"v1/user_role",
			"v1/user_role/*",
			"v1/user",
			"v1/user/*",
			"v1/selfuser", // Any user is allowed to use the login token to retrieve his/her own user info. temporarily given PERM_AUTHORIZATION for retrieving caller's user info
			"v1/api_key",
			"v1/api_key/*",
			"v1/selfapikey",
		},
		CONST_API_PWD_PROFILE: []string{
			"v1/password_profile",
			"v1/password_profile/*",
		},
		CONST_API_SYSTEM_CONFIG: []string{
			"v1/partner/ibm_sa_ep",
			"v1/partner/ibm_sa_config",
			"v1/file/config",
			"v1/system/config",
			"v2/system/config",
			"v1/system/license",
			"v1/system/summary",
			"v1/internal/system",
		},
		CONST_API_FED: []string{
			"v1/fed/join_token",
			"v1/fed/cluster/*/**",
			"v1/fed/view/*",
		},
		CONST_API_VULNERABILITY: []string{
			"v1/vulnerability/profile",
			"v1/vulnerability/profile/*",
		},
	}

	apiURIsPOST := map[int8][]string{
		CONST_API_NO_AUTH: []string{
			"v1/token_auth_server/*",
			"v1/fed/ping_internal",
			"v1/fed/joint_test_internal",
			"v1/auth",
			"v1/fed_auth",
			"v1/auth/*",
			"v1/eula",
		},
		CONST_API_DEBUG: []string{
			"v1/fed/promote",
			"v1/fed/join",
			"v1/fed/leave",
			"v1/fed/remove_internal",
			"v1/fed/command_internal",
			"v1/debug/controller/sync/*",
			"v1/controller/*/profiling",
			"v1/enforcer/*/profiling",
			"v1/file/config",
			"v1/csp/file/support",
			"v1/internal/alert",
		},
		CONST_API_RT_SCAN: []string{
			"v1/scan/workload/*",
			"v1/scan/host/*",
			"v1/scan/platform/platform",
			"v1/vulasset",
			"v1/assetvul",
		},
		CONST_API_REG_SCAN: []string{
			"v1/scan/registry/*/scan",
			"v1/scan/registry",
			"v2/scan/registry",
			"v1/scan/registry/*/test",
			"v2/scan/registry/*/test",
			"v1/scan/sigstore/root_of_trust",
			"v1/scan/sigstore/root_of_trust/*/verifier",
		},
		CONST_API_CICD_SCAN: []string{
			"v1/scan/result/repository",
			"v1/scan/repository",
		},
		CONST_API_GROUP: []string{
			"v1/group",
			"v1/file/group", // export group
			"v1/service",
		},
		CONST_API_RT_POLICIES: []string{
			"v1/workload/request/*",
			"v1/dlp/sensor",
			"v1/waf/sensor",
			"v1/file/dlp",
			"v1/file/dlp/config",
			"v1/file/waf",
			"v1/file/waf/config",
			"v1/system/request",
			"v1/sniffer",
			"v1/file/group/config", // for providing similar function as crd import but do not rely on crd webhook
		},
		CONST_API_ADM_CONTROL: []string{
			"v1/debug/admission/test",
			"v1/admission/rule",
			"v1/assess/admission/rule",
			"v1/file/admission",
			"v1/file/admission/config", // for providing similar function as crd import but do not rely on crd webhook
		},
		CONST_API_COMPLIANCE: []string{
			"v1/bench/host/*/docker",
			"v1/bench/host/*/kubernetes",
			"v1/file/compliance/profile",
			"v1/file/compliance/profile/config",
		},
		CONST_API_AUTHENTICATION: []string{
			"v1/server",
			"v1/debug/server/test",
		},
		CONST_API_AUTHORIZATION: []string{
			"v1/user_role",
			"v1/user",
			"v1/api_key",
			"v1/user/*/password",
		},
		CONST_API_PWD_PROFILE: []string{
			"v1/password_profile",
		},
		CONST_API_SYSTEM_CONFIG: []string{
			"v1/system/license/update",
			"v1/system/config/webhook",
			"v1/system/config/remote_repository",
		},
		CONST_API_IBMSA: []string{
			"v1/partner/ibm_sa/*/setup/*",
		},
		CONST_API_FED: []string{
			"v1/fed/demote",
			"v1/fed/deploy",
			"v1/fed/cluster/*/**",
			"v1/policy/rules/promote",
			"v1/admission/rule/promote",
		},
		CONST_API_VULNERABILITY: []string{
			"v1/vulnerability/profile/*/entry",
			"v1/file/vulnerability/profile",
			"v1/file/vulnerability/profile/config",
		},
	}

	apiURIsPATCH := map[int8][]string{
		CONST_API_NO_AUTH: []string{
			"v1/auth",
		},
		CONST_API_DEBUG: []string{
			"v1/fed/config",
		},
		CONST_API_RT_SCAN: []string{
			"v1/scan/config",
		},
		CONST_API_REG_SCAN: []string{
			"v1/scan/registry/*",
			"v2/scan/registry/*",
			"v1/scan/sigstore/root_of_trust/*",
			"v1/scan/sigstore/root_of_trust/*/verifier/*",
		},
		CONST_API_INFRA: []string{
			"v1/domain",
			"v1/domain/*",
		},
		CONST_API_NV_RESOURCE: []string{
			"v1/controller/*",
			"v1/enforcer/*",
		},
		CONST_API_GROUP: []string{
			"v1/group/*",
			"v1/service/config",
			"v1/service/config/network",
			"v1/service/config/profile",
		},
		CONST_API_RT_POLICIES: []string{
			"v1/workload/*",
			"v1/dlp/sensor/*",
			"v1/dlp/group/*",
			"v1/waf/sensor/*",
			"v1/waf/group/*",
			"v1/policy/rule",
			"v1/policy/rule/*",
			"v1/conversation_endpoint/*",
			"v1/process_profile/*",
			"v1/file_monitor/*",
			"v1/response/rule",
			"v1/response/rule/*",
			"v1/sniffer/stop/*",
		},
		CONST_API_ADM_CONTROL: []string{
			"v1/admission/state",
			"v1/admission/rule",
		},
		CONST_API_COMPLIANCE: []string{
			"v1/custom_check/*",
			"v1/compliance/profile/*",
			"v1/compliance/profile/*/entry/*",
		},
		CONST_API_AUTHENTICATION: []string{
			"v1/server/*",
			"v1/server/*/role/*",
			"v1/server/*/group/*",
			"v1/server/*/groups",
		},
		CONST_API_AUTHORIZATION: []string{
			"v1/user_role/*",
			"v1/user/*",
			"v1/user/*/role/*",
		},
		CONST_API_PWD_PROFILE: []string{
			"v1/password_profile/*",
		},
		CONST_API_SYSTEM_CONFIG: []string{
			"v1/system/config",
			"v2/system/config",
			"v1/system/config/webhook/*",
			"v1/system/config/remote_repository/*",
		},
		CONST_API_FED: []string{
			"v1/fed/cluster/*/**",
		},
		CONST_API_VULNERABILITY: []string{
			"v1/vulnerability/profile/*",
			"v1/vulnerability/profile/*/entry/*",
		},
	}

	apiURIsDELETE := map[int8][]string{
		CONST_API_NO_AUTH: []string{
			"v1/auth",
		},
		CONST_API_DEBUG: []string{
			"v1/fed_auth",
			"v1/conversation_endpoint/*",
			"v1/conversation",
			"v1/session",
			"v1/partner/ibm_sa/*/setup/*/*", // not supported by NV/IBMSA yet. Only for internal testing [20200831]
		},
		CONST_API_REG_SCAN: []string{
			"v1/scan/registry/*/scan",
			"v1/scan/registry/*",
			"v1/scan/registry/*/test",
			"v1/scan/sigstore/root_of_trust/*",
			"v1/scan/sigstore/root_of_trust/*/verifier/*",
		},
		CONST_API_GROUP: []string{
			"v1/group/*",
		},
		CONST_API_RT_POLICIES: []string{
			"v1/dlp/sensor/*",
			"v1/waf/sensor/*",
			"v1/policy/rule/*",
			"v1/policy/rule",
			"v1/conversation/*/*",
			"v1/response/rule/*",
			"v1/response/rule",
			"v1/sniffer/*",
		},
		CONST_API_ADM_CONTROL: []string{
			"v1/admission/rule/*",
			"v1/admission/rules",
		},
		CONST_API_COMPLIANCE: []string{
			"v1/compliance/profile/*/entry/*",
		},
		CONST_API_AUTHENTICATION: []string{
			"v1/server/*",
		},
		CONST_API_AUTHORIZATION: []string{
			"v1/user_role/*",
			"v1/user/*",
			"v1/api_key/*",
		},
		CONST_API_PWD_PROFILE: []string{
			"v1/password_profile/*",
		},
		CONST_API_SYSTEM_CONFIG: []string{
			"v1/system/license",
			"v1/system/config/webhook/*",
			"v1/system/config/remote_repository/*",
		},
		CONST_API_FED: []string{
			"v1/fed/cluster/*",
			"v1/fed/cluster/*/**",
		},
		CONST_API_VULNERABILITY: []string{
			"v1/vulnerability/profile/*/entry/*",
		},
	}

	verbApiURIsMappingData := map[string]map[int8][]string{
		"GET":    apiURIsGET,
		"POST":   apiURIsPOST,
		"PATCH":  apiURIsPATCH,
		"DELETE": apiURIsDELETE,
	}

	for verb, apiURIsMappingData := range verbApiURIsMappingData {
		for apiID, uris := range apiURIsMappingData {
			for _, uri := range uris {
				url := fmt.Sprintf("https://10.1.1.1/%s", uri)
				if r, err := http.NewRequest(verb, url, nil); err == nil {
					apiCategoryID, requiredPermissions := getRequiredPermissions(r)
					if apiCategoryID != apiID {
						t.Errorf("got wrong api category id for url(%s): expected=%v, got=%+v", url, apiID, apiCategoryID)
					} else if p, ok := apiPermissions[apiID]; !ok || p != requiredPermissions {
						t.Errorf("got wrong required permissions for url(%s): expected=%v, got=%+v", url, p, requiredPermissions)
					}
				} else {
					t.Errorf("new request failed url(%s): err=%v", url, err)
				}
			}
		}
	}

	postTest()
}

package access

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/neuvector/neuvector/controller/api"
	"github.com/neuvector/neuvector/share"
	"github.com/neuvector/neuvector/share/utils"
)

const (
	CONST_PERM_SUPPORT_GLOBAL = 0x1
	CONST_PERM_SUPPORT_DOMAIN = 0x2
	CONST_PERM_SUPPORT_BOTH   = 0x3 // CONST_PERM_SUPPORT_GLOBAL + CONST_PERM_SUPPORT_DOMAIN
)

const (
	CONST_VISIBLE_USER_ROLE   = iota // roles that can be associated with global domain
	CONST_VISIBLE_DOMAIN_ROLE        // domaon roles & mappable group domain roles are the same set
	CONST_MAPPABLE_SERVER_DEFAULT_ROLE
)

// apiCategoryID
const (
	CONST_API_UNKNOWN = iota
	CONST_API_UNSUPPORTED
	CONST_API_SKIP
	CONST_API_NO_AUTH
	CONST_API_DEBUG // i.e. for admin only
	CONST_API_RT_SCAN
	CONST_API_REG_SCAN
	CONST_API_CICD_SCAN
	CONST_API_CLOUD
	CONST_API_INFRA
	CONST_API_NV_RESOURCE
	CONST_API_WORKLOAD
	CONST_API_GROUP
	CONST_API_RT_POLICIES
	CONST_API_ADM_CONTROL
	CONST_API_COMPLIANCE
	CONST_API_AUDIT_EVENTS
	CONST_API_SECURITY_EVENTS
	CONST_API_EVENTS
	CONST_API_AUTHENTICATION
	CONST_API_AUTHORIZATION
	CONST_API_SYSTEM_CONFIG
	CONST_API_IBMSA
	CONST_API_FED
	CONST_API_PWD_PROFILE   // i.e. for password profile
	CONST_API_VULNERABILITY // i.e. for vulnerability profile
)

// apiCategoryID to permissions mapping
var apiPermissions map[int8]uint64 = map[int8]uint64{ // key is apiCategoryID
	CONST_API_UNKNOWN:         0,
	CONST_API_NO_AUTH:         0,
	CONST_API_DEBUG:           share.PERMS_CLUSTER_WRITE,
	CONST_API_RT_SCAN:         share.PERMS_RUNTIME_SCAN,
	CONST_API_REG_SCAN:        share.PERM_REG_SCAN,
	CONST_API_CICD_SCAN:       share.PERM_CICD_SCAN,
	CONST_API_CLOUD:           share.PERM_CLOUD,
	CONST_API_INFRA:           share.PERM_INFRA_BASIC,
	CONST_API_NV_RESOURCE:     share.PERM_NV_RESOURCE,
	CONST_API_WORKLOAD:        share.PERM_WORKLOAD_BASIC,
	CONST_API_GROUP:           share.PERM_GROUP_BASIC,
	CONST_API_RT_POLICIES:     share.PERMS_RUNTIME_POLICIES,
	CONST_API_ADM_CONTROL:     share.PERM_ADM_CONTROL,
	CONST_API_COMPLIANCE:      share.PERMS_COMPLIANCE,
	CONST_API_AUDIT_EVENTS:    share.PERM_AUDIT_EVENTS,
	CONST_API_SECURITY_EVENTS: share.PERMS_SECURITY_EVENTS,
	CONST_API_EVENTS:          share.PERM_EVENTS,
	CONST_API_AUTHENTICATION:  share.PERM_AUTHENTICATION,
	CONST_API_AUTHORIZATION:   share.PERM_AUTHORIZATION,
	CONST_API_SYSTEM_CONFIG:   share.PERM_SYSTEM_CONFIG,
	CONST_API_IBMSA:           share.PERM_IBMSA,
	CONST_API_FED:             share.PERM_FED,
	CONST_API_PWD_PROFILE:     share.PERMS_PWD_PROFILE,  // i.e. for password profile
	CONST_API_VULNERABILITY:   share.PERM_VULNERABILITY, // i.e. for vulnerability profile
}

// key is permission id that is visible to the world. Regarding to the value,
// 1. if len(value.ComplexPermits) == 0, value is the effective internal permission used by controller
// 2. if len(value.ComplexPermits) > 0, value.ComplexPermits has the effective internal permissions used by controller
var PermissionOptions = []*api.RESTRolePermitOptionInternal{ // basic permission can only be contained by other permissions
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_SYSTEM_CONFIG_ID,
		Value:          share.PERM_SYSTEM_CONFIG,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_IBMSA_ID,
		Value:          share.PERM_IBMSA,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_FED_ID,
		Value:          share.PERM_FED,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_NV_RESOURCE_ID,
		Value:          share.PERM_NV_RESOURCE,
		SupportScope:   CONST_PERM_SUPPORT_BOTH,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERMS_RUNTIME_SCAN_ID,
		Value:          share.PERMS_RUNTIME_SCAN,
		SupportScope:   CONST_PERM_SUPPORT_BOTH,
		ReadSupported:  true,
		WriteSupported: true,
		ComplexPermits: []*api.RESTRolePermitOptionInternal{
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_RUNTIME_SCAN_BASIC_ID,
				Value:          share.PERM_RUNTIME_SCAN_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_WORKLOAD_BASIC_ID,
				Value:          share.PERM_WORKLOAD_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_INFRA_BASIC_ID,
				Value:          share.PERM_INFRA_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
				ReadSupported:  true,
				WriteSupported: true,
			},
		},
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_REG_SCAN_ID,
		Value:          share.PERM_REG_SCAN,
		SupportScope:   CONST_PERM_SUPPORT_BOTH,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_CICD_SCAN_ID,
		Value:          share.PERM_CICD_SCAN,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:           share.PERM_CLOUD_ID,
		Value:        share.PERM_CLOUD,
		SupportScope: CONST_PERM_SUPPORT_GLOBAL,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERMS_RUNTIME_POLICIES_ID,
		Value:          share.PERMS_RUNTIME_POLICIES,
		SupportScope:   CONST_PERM_SUPPORT_BOTH,
		ReadSupported:  true,
		WriteSupported: true,
		ComplexPermits: []*api.RESTRolePermitOptionInternal{
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_GROUP_BASIC_ID,
				Value:          share.PERM_GROUP_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_NETWORK_POLICY_BASIC_ID,
				Value:          share.PERM_NETWORK_POLICY_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_SYSTEM_POLICY_BASIC_ID,
				Value:          share.PERM_SYSTEM_POLICY_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_WORKLOAD_BASIC_ID,
				Value:          share.PERM_WORKLOAD_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
		},
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_ADM_CONTROL_ID,
		Value:          share.PERM_ADM_CONTROL,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERMS_COMPLIANCE_ID,
		Value:          share.PERMS_COMPLIANCE,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
		ComplexPermits: []*api.RESTRolePermitOptionInternal{
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_COMPLIANCE_BASIC_ID,
				Value:          share.PERM_COMPLIANCE_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_WORKLOAD_BASIC_ID,
				Value:          share.PERM_WORKLOAD_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_BOTH,
				ReadSupported:  true,
				WriteSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:             share.PERM_INFRA_BASIC_ID,
				Value:          share.PERM_INFRA_BASIC,
				SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
				ReadSupported:  true,
				WriteSupported: true,
			},
		},
	},
	&api.RESTRolePermitOptionInternal{
		ID:            share.PERM_AUDIT_EVENTS_ID,
		Value:         share.PERM_AUDIT_EVENTS,
		SupportScope:  CONST_PERM_SUPPORT_BOTH,
		ReadSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:            share.PERMS_SECURITY_EVENTS_ID,
		Value:         share.PERMS_SECURITY_EVENTS,
		SupportScope:  CONST_PERM_SUPPORT_BOTH,
		ReadSupported: true,
		ComplexPermits: []*api.RESTRolePermitOptionInternal{
			&api.RESTRolePermitOptionInternal{
				ID:            share.PERM_SECURITY_EVENTS_BASIC_ID,
				Value:         share.PERM_SECURITY_EVENTS_BASIC,
				SupportScope:  CONST_PERM_SUPPORT_BOTH,
				ReadSupported: true,
			},
			&api.RESTRolePermitOptionInternal{
				ID:            share.PERM_WORKLOAD_BASIC_ID,
				Value:         share.PERM_WORKLOAD_BASIC,
				SupportScope:  CONST_PERM_SUPPORT_BOTH,
				ReadSupported: true,
			},
		},
	},
	&api.RESTRolePermitOptionInternal{
		ID:            share.PERM_EVENTS_ID,
		Value:         share.PERM_EVENTS,
		SupportScope:  CONST_PERM_SUPPORT_BOTH,
		ReadSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_AUTHENTICATION_ID,
		Value:          share.PERM_AUTHENTICATION,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_AUTHORIZATION_ID,
		Value:          share.PERM_AUTHORIZATION,
		SupportScope:   CONST_PERM_SUPPORT_BOTH,
		ReadSupported:  true,
		WriteSupported: true,
	},
	&api.RESTRolePermitOptionInternal{
		ID:             share.PERM_VULNERABILITY_ID,
		Value:          share.PERM_VULNERABILITY,
		SupportScope:   CONST_PERM_SUPPORT_GLOBAL,
		ReadSupported:  true,
		WriteSupported: true,
	},
}

var HiddenPermissions = utils.NewSet(share.PERM_IBMSA_ID, share.PERM_FED_ID, share.PERM_CLOUD_ID, share.PERM_NV_RESOURCE_ID)
var hiddenRoles = utils.NewSet(api.UserRoleFedAdmin, api.UserRoleFedReader, api.UserRoleIBMSA, api.UserRoleImportStatus)
var visibleRoles = utils.NewSet(api.UserRoleAdmin, api.UserRoleReader, api.UserRoleNone, api.UserRoleCIOps)               // fedAdmin/fedReader roles are added when this is master cluster
var mappableServerDefaultRoles = utils.NewSet(api.UserRoleAdmin, api.UserRoleReader, api.UserRoleCIOps, api.UserRoleNone) // for default role of server
var mappableDomainRoles = utils.NewSet(api.UserRoleAdmin, api.UserRoleReader, api.UserRoleCIOps)                          // for groups' mapped domain roles of server
var allRoles = map[string]*share.CLUSUserRoleInternal{                                                                    // changed from CustomRolesActions, key is role name, value is role permission
	api.UserRoleFedAdmin: &share.CLUSUserRoleInternal{
		Name:         api.UserRoleFedAdmin,
		Comment:      "Federated Administrator role",
		Reserved:     true,
		ReadPermits:  share.PERMS_FED_READ,
		WritePermits: share.PERMS_FED_WRITE,
	},
	api.UserRoleFedReader: &share.CLUSUserRoleInternal{
		Name:        api.UserRoleFedReader,
		Comment:     "Federated View role",
		Reserved:    true,
		ReadPermits: share.PERMS_FED_READ,
	},
	api.UserRoleAdmin: &share.CLUSUserRoleInternal{
		Name:         api.UserRoleAdmin,
		Comment:      "Global Administrator role",
		Reserved:     true,
		ReadPermits:  share.PERMS_CLUSTER_READ,
		WritePermits: share.PERMS_CLUSTER_WRITE,
	},
	api.UserRoleReader: &share.CLUSUserRoleInternal{
		Name:        api.UserRoleReader,
		Comment:     "Global View role",
		Reserved:    true,
		ReadPermits: share.PERMS_CLUSTER_READ,
	},
	api.UserRoleNone: &share.CLUSUserRoleInternal{
		Name:        api.UserRoleNone,
		Comment:     "None",
		Reserved:    true,
		ReadPermits: 0,
	},
	api.UserRoleCIOps: &share.CLUSUserRoleInternal{
		Name:         api.UserRoleCIOps,
		Comment:      "CI Integration role",
		Reserved:     true,
		WritePermits: share.PERM_CICD_SCAN,
	},
	api.UserRoleIBMSA: &share.CLUSUserRoleInternal{
		Name:         api.UserRoleIBMSA,
		Comment:      "for IBM Security Advisor",
		Reserved:     true,
		ReadPermits:  share.PERM_IBMSA,
		WritePermits: share.PERM_IBMSA,
	},
	api.UserRoleImportStatus: &share.CLUSUserRoleInternal{
		Name:        api.UserRoleImportStatus,
		Comment:     "for reading import status",
		Reserved:    true,
		ReadPermits: share.PERM_SYSTEM_CONFIG,
	},
}

var rolesMutex sync.RWMutex

func clusUserRoleToREST(name string, r *share.CLUSUserRoleInternal) *api.RESTUserRole {
	var permissions []*api.RESTRolePermission
	if name == api.UserRoleFedAdmin || name == api.UserRoleFedReader || name == api.UserRoleAdmin || name == api.UserRoleReader {
		permissions = make([]*api.RESTRolePermission, 0)
	} else {
		permissions = make([]*api.RESTRolePermission, 0, len(PermissionOptions))
		for _, option := range PermissionOptions {
			if HiddenPermissions.Contains(option.ID) {
				continue // do not return hidden permissions to rest client
			}
			permission := &api.RESTRolePermission{
				ID: option.ID,
			}
			if option.ReadSupported && ((r.ReadPermits & option.Value) == option.Value) {
				permission.Read = true
			}
			if option.WriteSupported && ((r.WritePermits & option.Value) == option.Value) {
				permission.Write = true
			}
			if permission.Read || permission.Write {
				permissions = append(permissions, permission)
			}
		}
	}
	role := &api.RESTUserRole{
		Name:        name,
		Comment:     r.Comment,
		Reserved:    r.Reserved,
		Permissions: permissions,
	}

	return role
}

func getRolePermitValues(roleName, domain string) (uint64, uint64) {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	if role, ok := allRoles[roleName]; ok {
		if roleName == api.UserRoleIBMSA || roleName == api.UserRoleImportStatus {
			if domain == AccessDomainGlobal {
				return role.ReadPermits, role.WritePermits
			} else {
				return 0, 0 // ibm_sa permission is not effective in domain role
			}
		} else {
			if domain == AccessDomainGlobal {
				return role.ReadPermits & share.PERMS_FED_READ, role.WritePermits & share.PERMS_FED_WRITE // filter out unsupported permission on global role
			} else {
				return role.ReadPermits & share.PERMS_DOMAIN_READ, role.WritePermits & share.PERMS_DOMAIN_WRITE // filter out unsupported permission on domain role
			}
		}
	}
	return 0, 0
}

func getRestRolePermitValues(roleName, domain string) []*api.RESTRolePermission {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	//rolePermits := map[string]*api.RESTRolePermission	// key is permission id
	var pList []*api.RESTRolePermission
	if role, ok := allRoles[roleName]; ok {
		pList = make([]*api.RESTRolePermission, 0, len(PermissionOptions))
		readPermits, writePermits := role.ReadPermits, role.WritePermits
		if domain == AccessDomainGlobal {
			readPermits &= share.PERMS_FED_READ
			writePermits &= share.PERMS_FED_WRITE
		} else {
			readPermits &= share.PERMS_DOMAIN_READ
			writePermits &= share.PERMS_DOMAIN_WRITE
		}
		for _, option := range PermissionOptions {
			// need to check fed/nv_resource permissions as well
			var read, write bool
			optionReadValue, optionWriteValue := option.Value, option.Value
			if domain == AccessDomainGlobal {
				optionReadValue &= share.PERMS_FED_READ
				optionWriteValue &= share.PERMS_FED_WRITE
			} else {
				optionReadValue &= share.PERMS_DOMAIN_READ
				optionWriteValue &= share.PERMS_DOMAIN_WRITE
			}
			if optionReadValue != 0 && option.ReadSupported && ((readPermits & optionReadValue) == optionReadValue) {
				read = true
			}
			if optionWriteValue != 0 && option.WriteSupported && ((writePermits & optionWriteValue) == optionWriteValue) {
				write = true
			}
			if read || write {
				p := &api.RESTRolePermission{
					ID:    option.ID,
					Read:  read,
					Write: write,
				}
				pList = append(pList, p)
			}
		}
	} else {
		pList = make([]*api.RESTRolePermission, 0)
	}

	return pList
}

func GetDomainPermissions(globalRole string, roleDomains map[string][]string) ([]*api.RESTRolePermission, map[string][]*api.RESTRolePermission, error) {
	var globalPermits []*api.RESTRolePermission
	domainPermits := make(map[string][]*api.RESTRolePermission)
	if globalRole != api.UserRoleNone {
		globalPermits = getRestRolePermitValues(globalRole, AccessDomainGlobal)
	} else {
		globalPermits = make([]*api.RESTRolePermission, 0)
	}
	count := len(globalPermits)
	for role, domains := range roleDomains {
		if role == api.UserRoleNone {
			continue
		}
		var domainPicked string
		for _, domain := range domains {
			if domain != AccessDomainGlobal {
				domainPicked = domain
				break
			}
		}
		if domainPicked != "" {
			permissions := getRestRolePermitValues(role, domainPicked)
			for _, domain := range domains {
				domainPermits[domain] = permissions
			}
			count += len(permissions)
		}
	}
	if count == 0 {
		return nil, nil, fmt.Errorf("This user has no permission enabled!")
	}

	return globalPermits, domainPermits, nil
}

type UriApiNode struct {
	childNodes    map[string]*UriApiNode
	apiCategoryID int8 // this uri(so far from root node to this node) belongs to which API category. -1 means there is no API handler for this URI
}

var uriRequiredPermitsMappings map[string]*UriApiNode // key is method

func getRequiredPermissions(r *http.Request) (int8, uint64) {
	if r == nil {
		return 0, 0
	}

	var apiCategoryID int8

	u, err := url.Parse(strings.ToLower(r.URL.String()))
	if err == nil {
		// u.Path is like "/v1/log/event"
		ss := strings.Split(u.Path, "/") // ss is like {"", "v1", "log", "event"}
		ssEndIndex := len(ss) - 2        // it's 2 because we iterate from ss[1]
		// ignore leading "" in ss
		if parentNode, ok := uriRequiredPermitsMappings[r.Method]; ok {
			for idx, s := range ss[1:] {
				if node, ok := parentNode.childNodes["**"]; ok {
					// forward requests(for multi-clusters) reach here
					apiCategoryID = node.apiCategoryID
					break
				} else if node, ok := parentNode.childNodes["*"]; ok {
					if idx == ssEndIndex {
						apiCategoryID = node.apiCategoryID
						break
					} else {
						parentNode = node
					}
				} else if node, ok := parentNode.childNodes[s]; ok {
					if idx == ssEndIndex {
						apiCategoryID = node.apiCategoryID
						break
					} else {
						parentNode = node
					}
				} else {
					break
				}
			}
		}
	}
	requiredPermissions, _ := apiPermissions[apiCategoryID]

	return apiCategoryID, requiredPermissions
}

func parseForRequiredPermits(ssUri []string, parentNode *UriApiNode, apiID int8) bool { // ssUri is like {"v1", "log", "event"} for GET("/v1/log/event"). return true means caller is leaf node.
	if len(ssUri) == 0 {
		return true
	}
	if currentNode, ok := parentNode.childNodes[ssUri[0]]; !ok || currentNode == nil {
		currentNode = &UriApiNode{
			childNodes:    make(map[string]*UriApiNode),
			apiCategoryID: CONST_API_UNSUPPORTED,
		}
		if amILeafNode := parseForRequiredPermits(ssUri[1:], currentNode, apiID); amILeafNode { // advance to next part in URI
			currentNode.apiCategoryID = apiID
		}
		parentNode.childNodes[ssUri[0]] = currentNode
	} else {
		if amILeafNode := parseForRequiredPermits(ssUri[1:], currentNode, apiID); amILeafNode {
			currentNode.apiCategoryID = apiID
		}
	}
	return false
}

/*
func dumpApiUriParts(verb, parentURI string, nodes map[string]*UriApiNode) { // ssUri is like {"v1", "log", "event"} for GET("/v1/log/event"). return true means caller is leaf node.
	if len(nodes) == 0 {
		return
	}
	for part, node := range nodes {
		if node != nil {
			nodeURI := fmt.Sprintf("%s/%s", parentURI, part)
			dumpApiUriParts(verb, nodeURI, node.childNodes)
			fmt.Printf("[dump] --------------> verb=%s, nodeURI=%s, apiID=%d\n", verb, nodeURI, node.apiCategoryID)
		}
	}
	return
}
*/
func CompileUriPermitsMapping() {
	if uriRequiredPermitsMappings == nil {
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
				// "scanasset5", // TODO: local dev only
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
				"v1/group/*/stats",
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
				// "scanasset5",     // TODO: local dev only
				// "scanassetview1", // TODO: local dev only
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

		uriRequiredPermitsMappings = make(map[string]*UriApiNode, 4)

		verbApiURIsMappingData := map[string]map[int8][]string{
			"GET":    apiURIsGET,
			"POST":   apiURIsPOST,
			"PATCH":  apiURIsPATCH,
			"DELETE": apiURIsDELETE,
		}
		for verb, apiURIsMappingData := range verbApiURIsMappingData {
			currentNode := &UriApiNode{
				childNodes:    make(map[string]*UriApiNode, len(apiURIsMappingData)),
				apiCategoryID: CONST_API_UNSUPPORTED,
			}
			for apiID, uris := range apiURIsMappingData {
				for _, uri := range uris {
					ss := strings.Split(uri, "/")
					// ss is like {"v1", "log", "event"} for GET("/v1/log/event")
					parseForRequiredPermits(ss, currentNode, apiID)
				}
			}
			uriRequiredPermitsMappings[verb] = currentNode
		}

		/* [debug] dump uriRequiredPermitsMappings
		for verb, currentNode := range uriRequiredPermitsMappings {
			dumpApiUriParts(verb, "", currentNode.childNodes)
		}
		*/
	}
}

func IsValidRole(role string, usage int) bool {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	switch usage {
	case CONST_VISIBLE_USER_ROLE:
		return visibleRoles.Contains(role)
	case CONST_VISIBLE_DOMAIN_ROLE: // domain roles & mappable group domain roles are the same set
		return mappableDomainRoles.Contains(role)
	case CONST_MAPPABLE_SERVER_DEFAULT_ROLE:
		return mappableServerDefaultRoles.Contains(role)
	}
	return false
}

func GetValidRoles(usage int) []string {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	var rolesSet utils.Set
	switch usage {
	case CONST_VISIBLE_USER_ROLE:
		rolesSet = visibleRoles
	case CONST_VISIBLE_DOMAIN_ROLE: // domain roles & mappable group domain roles are the same set
		rolesSet = mappableDomainRoles
	case CONST_MAPPABLE_SERVER_DEFAULT_ROLE:
		rolesSet = mappableServerDefaultRoles
	default:
		return nil
	}
	rolesSet = rolesSet.Clone()
	// UI requires the reserved roles in fixed leading positions, so sor it here
	roles := make([]string, 0, rolesSet.Cardinality())
	reservedRoles := []string{api.UserRoleFedAdmin, api.UserRoleFedReader, api.UserRoleAdmin, api.UserRoleReader, api.UserRoleCIOps, api.UserRoleNone}
	for _, reservedRole := range reservedRoles {
		if rolesSet.Contains(reservedRole) {
			roles = append(roles, reservedRole)
			rolesSet.Remove(reservedRole)
		}
	}
	remainRoles := rolesSet.ToStringSlice()
	sort.Slice(remainRoles, func(i, j int) bool {
		return remainRoles[i] < remainRoles[j]
	})
	roles = append(roles, remainRoles...)
	return roles
}

func AddRole(name string, role *share.CLUSUserRoleInternal) {
	rolesMutex.Lock()
	defer rolesMutex.Unlock()

	visibleRoles.Add(name)
	mappableServerDefaultRoles.Add(name)
	mappableDomainRoles.Add(name)
	if r, ok := allRoles[name]; ok {
		r.Comment = role.Comment
		r.ReadPermits = role.ReadPermits
		r.WritePermits = role.WritePermits
	} else {
		allRoles[name] = role
	}
}

func DeleteRole(name string) {
	rolesMutex.Lock()
	defer rolesMutex.Unlock()

	if role, ok := allRoles[name]; ok {
		if !role.Reserved {
			visibleRoles.Remove(name)
			mappableServerDefaultRoles.Remove(name)
			mappableDomainRoles.Remove(name)
			delete(allRoles, name)
		}
	}
}

func UpdateUserRoleForFedRoleChange(fedRole string) {
	rolesMutex.Lock()
	defer rolesMutex.Unlock()

	roles := []string{api.UserRoleFedAdmin, api.UserRoleFedReader}
	for _, role := range roles {
		if fedRole == api.FedRoleMaster {
			visibleRoles.Add(role)
			hiddenRoles.Remove(role)
		} else {
			visibleRoles.Remove(role)
			hiddenRoles.Add(role)
		}
	}
}

func GetRoleList() []*api.RESTUserRole {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	names := GetValidRoles(CONST_VISIBLE_USER_ROLE)
	roles := make([]*api.RESTUserRole, 0, len(names))
	for _, name := range names {
		if name == api.UserRoleNone {
			continue
		}
		if role, ok := allRoles[name]; ok {
			roleRest := clusUserRoleToREST(name, role)
			if roleRest != nil {
				roles = append(roles, roleRest)
			}
		}
	}

	return roles
}

func GetRoleDetails(name string) *api.RESTUserRole {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	if visibleRoles.Contains(name) {
		if role, ok := allRoles[name]; ok {
			return clusUserRoleToREST(name, role)
		}
	}
	return nil
}

func GetReservedRoleNames() utils.Set {
	rolesMutex.RLock()
	defer rolesMutex.RUnlock()

	names := utils.NewSet()
	for _, role := range allRoles {
		if role.Reserved {
			names.Add(role.Name)
		}
	}
	return names
}

//--------
type AccessOP string

const (
	AccessOPRead  AccessOP = "read"
	AccessOPWrite          = "write"
)

const AccessDomainGlobal = ""

type DomainRole map[string]string // domain -> role

// check if for global domain it has the specified permissions. required permission value being 0 means we don't care about that permission
func (drs DomainRole) hasGlobalPermissions(readPermsRequired, writePermsRequired uint64) bool {
	readPermsRequired &= share.PERMS_FED_READ
	writePermsRequired &= share.PERMS_FED_WRITE
	if role, ok := drs[AccessDomainGlobal]; ok {
		readPermits, writePermits := getRolePermitValues(role, AccessDomainGlobal)
		if (readPermsRequired == (readPermits & readPermsRequired)) && (writePermsRequired == (writePermits & writePermsRequired)) {
			return true
		}
	}
	return false
}

// check if it has the specified permissions in any domain. required permission value being 0 means we don't care about that permission
func (drs DomainRole) hasPermissions(_readPermitsRequired, _writePermsRequired uint64) bool {
	var readPermitsRequired, writePermsRequired uint64
	for domain, role := range drs {
		if domain == AccessDomainGlobal {
			readPermitsRequired = _readPermitsRequired & share.PERMS_FED_READ
			writePermsRequired = _writePermsRequired & share.PERMS_FED_WRITE
		} else {
			readPermitsRequired = _readPermitsRequired & share.PERMS_DOMAIN_READ
			writePermsRequired = _writePermsRequired & share.PERMS_DOMAIN_WRITE
		}
		readPermits, writePermits := getRolePermitValues(role, domain)
		if (readPermitsRequired == (readPermits & readPermitsRequired)) && (writePermsRequired == (writePermits & writePermsRequired)) {
			return true
		}
	}
	return false
}

type AccessControl struct {
	op     AccessOP
	roles  DomainRole // domain -> role
	wRoles DomainRole // special domain(containing wildcard char) -> role

	// the API's category id where this AccessControl is created from. It's CONST_API_SKIP if this object is created from NewFedAdminAccessControl()/NewAdminAccessControl()/NewReaderAccessControl()
	apiCategoryID int8
	// required permissions to check for calling a REST API, 0 means access not allowed
	requiredPermissions uint64

	// permissions to boost for a REST API. Rarely used
	boostPermissions uint64
}

func NewAdminAccessControl() *AccessControl {
	return &AccessControl{
		op:            AccessOPWrite,
		roles:         map[string]string{AccessDomainGlobal: api.UserRoleAdmin},
		wRoles:        map[string]string{},
		apiCategoryID: CONST_API_SKIP,
	}
}

// be careful when using this function because it returns a very powerful access control object
func NewFedAdminAccessControl() *AccessControl {
	return &AccessControl{
		op:            AccessOPWrite,
		roles:         map[string]string{AccessDomainGlobal: api.UserRoleFedAdmin},
		wRoles:        map[string]string{},
		apiCategoryID: CONST_API_SKIP,
	}
}

func NewReaderAccessControl() *AccessControl {
	return &AccessControl{
		op:            AccessOPRead,
		roles:         map[string]string{AccessDomainGlobal: api.UserRoleReader},
		wRoles:        map[string]string{},
		apiCategoryID: CONST_API_SKIP,
	}
}

func NewAccessControl(r *http.Request, op AccessOP, roles DomainRole) *AccessControl {
	wRoles := map[string]string{}
	for domain, role := range roles {
		if strings.Contains(domain, "*") {
			wRoles[domain] = role
		}
	}
	acc := &AccessControl{
		op:     op,
		roles:  roles,
		wRoles: wRoles,
	}
	acc.apiCategoryID, acc.requiredPermissions = getRequiredPermissions(r)
	/*
		if acc.requiredPermissions == 0 {
			log.WithFields(log.Fields{"method": r.Method, "url": r.URL.String(), "apiCategoryID": acc.apiCategoryID}).Warn("no permission required!")
		}
	*/

	return acc
}

// generate a new access control object that is the same as the calling object except the op is different
func (acc *AccessControl) NewWithOp(op AccessOP) *AccessControl {
	return &AccessControl{
		op:                  op,
		roles:               acc.roles,
		wRoles:              acc.wRoles,
		apiCategoryID:       acc.apiCategoryID,
		requiredPermissions: acc.requiredPermissions,
	}
}

// now we use API-level permission. So it's rare that an API needs to boost permissions for the caller
func (acc *AccessControl) BoostPermissions(toBoost uint64) *AccessControl {
	return &AccessControl{
		op:                  acc.op,
		roles:               acc.roles,
		wRoles:              acc.wRoles,
		apiCategoryID:       acc.apiCategoryID,
		requiredPermissions: acc.requiredPermissions,
		boostPermissions:    toBoost,
	}
}

func (acc *AccessControl) isDomainRoleAllowedToAccess(role, domain string, readPermitsRequired, writePermitsRequired uint64, accNotFromCaller bool) bool {
	readPermits, writePermits := getRolePermitValues(role, domain)
	if domain != AccessDomainGlobal || role != "" {
		// Boost permissions only when the caller's global role is not None. Otherwise namespace user could be boosted to see other domain's objects.
		// The purpose of permissions boost is to allow caller to see other types of objects, not to see other domain's objects.
		readPermits |= acc.boostPermissions
		writePermits |= acc.boostPermissions
	}
	if acc.op == AccessOPRead {
		if (readPermitsRequired != 0 || accNotFromCaller) && (readPermitsRequired == (readPermits & readPermitsRequired)) {
			return true
		}
	} else if acc.op == AccessOPWrite {
		if (writePermitsRequired != 0 || accNotFromCaller) && (writePermitsRequired == (writePermits & writePermitsRequired)) {
			return true
		}
	}

	return false
}

// The domain-param containing wildcard char or not (except for global domain),
// 	if there is any entry(domain, role) in acc.roles/acc.wroles that the entry.domain is superset of domain-param & the entry.role has the required permissions, it's allowed.
// Here 'superset' means string matching using regex.
// For global domain, even user has permission on "*" namespaces only, it's still a namespace user and cannot access global-only objects
// See TestWildcardDomainAccess*() & TestWildcardOwnAccess*() in access_test.go about the examples
func (acc *AccessControl) isOneAccessAllowed(domain string, readPermitsRequired, writePermitsRequired uint64) bool {
	// domain argument may contain wildcard character
	accNotFromCaller := false
	if acc.apiCategoryID == CONST_API_SKIP {
		accNotFromCaller = true
	}

	// eg. return []string{AccessAllAsReader}, nil
	if domain == share.AccessAllAsReader {
		// the resource can be read by global/namespace users with specific permissions from its global/domain role, only global user with specific permissions can write
		for d, role := range acc.roles {
			readPermits, writePermits := getRolePermitValues(role, domain)
			readPermits |= acc.boostPermissions
			writePermits |= acc.boostPermissions
			if d == AccessDomainGlobal {
				// Global user follow role/permission definition
				if acc.op == AccessOPRead {
					if (readPermitsRequired != 0 || accNotFromCaller) && (readPermitsRequired == (readPermits & readPermitsRequired)) {
						return true
					}
				} else if acc.op == AccessOPWrite {
					if (writePermitsRequired != 0 || accNotFromCaller) && (writePermitsRequired == (writePermits & writePermitsRequired)) {
						return true
					}
				}
			} else {
				// Domain user can read it if it has the required read permission for any domain
				if acc.op == AccessOPRead && (readPermitsRequired != 0 || accNotFromCaller) && (readPermitsRequired == (readPermits & readPermitsRequired)) {
					return true
				}
			}
		}
		return false
	}

	// keys in acc.roles may contain wildcard character as well
	// 1. acc.roles contains a key/value for the domain. The value is used to do role/permission matching
	// 2. acc.roles does not contain a key/value for the domain. skip
	if role, ok := acc.roles[domain]; ok {
		if acc.isDomainRoleAllowedToAccess(role, domain, readPermitsRequired, writePermitsRequired, accNotFromCaller) {
			return true
		}
	}

	// keys(user's domains) in acc.wRoles contain wildcard character. iterate thru all keys in acc.wRoles to see if the param domain is a subset of the key
	// 1. domain(of resource object) is subset of a key. Use the key's value to do role/permission matching
	// 2. domain(of resource object) is not subset of a key. skip this key and continue on next key
	for wDomain, role := range acc.wRoles {
		if domain != AccessDomainGlobal { // { role -> "*" namespace } is role to namespace mapping. { role -> "*" } still cannot cover global namespace("") !!
			if share.EqualMatch(wDomain, domain) { // if param domain matches a wildcard domain
				if acc.isDomainRoleAllowedToAccess(role, domain, readPermitsRequired, writePermitsRequired, accNotFromCaller) {
					return true
				}
			}
		}
	}

	return false
}

// Return true if one of domains is allowed to access
func (acc *AccessControl) isAccessAllowed(domains []string, readPermitsRequired, writePermitsRequired uint64) bool {
	for _, domain := range domains {
		if acc.isOneAccessAllowed(domain, readPermitsRequired, writePermitsRequired) {
			return true
		}
	}

	return false
}

// Return false if one of domains is not allowed to access
func (acc *AccessControl) isOwnAllowed(domains []string, readPermitsRequired, writePermitsRequired uint64) bool {
	for _, domain := range domains {
		if !acc.isOneAccessAllowed(domain, readPermitsRequired, writePermitsRequired) {
			return false
		}
	}
	return true
}

// returns true only when the access control object is created for user whose global role has the same permissions as fedAdmin role for read/write
func (acc *AccessControl) IsFedAdmin() bool {
	return acc.roles.hasGlobalPermissions(share.PERMS_FED_READ, share.PERMS_FED_WRITE)
}

// returns true only when the access control object is created for user whose global role has the same permissions as fedReader role for read
func (acc *AccessControl) IsFedReader() bool {
	var readPermsRequired uint64 = share.PERMS_FED_READ
	if role, ok := acc.roles[AccessDomainGlobal]; ok {
		readPermits, writePermits := getRolePermitValues(role, AccessDomainGlobal)
		if (readPermsRequired == (readPermits & readPermsRequired)) && (writePermits == 0) {
			return true
		}
	}
	return false
}

// returns true only when the access control object is created for user whose global role has the specified read/write permissions
func (acc *AccessControl) HasGlobalPermissions(readPermitsRequired, writePermsRequired uint64) bool {
	return acc.roles.hasGlobalPermissions(readPermitsRequired, writePermsRequired)
}

// returns true when the access control object is created for user whose role on any domain/global has the specified read/write permissions
func (acc *AccessControl) HasRequiredPermissions() bool {
	// check if required permissions for calling the API is met
	if acc.apiCategoryID == CONST_API_SKIP {
		// this acc object is not generated from caller's token. always return true
		return true
	} else if acc.apiCategoryID != CONST_API_UNSUPPORTED && acc.apiCategoryID != CONST_API_UNKNOWN {
		if acc.requiredPermissions != 0 {
			if acc.op == AccessOPWrite {
				return acc.roles.hasPermissions(0, acc.requiredPermissions)
			} else {
				return acc.roles.hasPermissions(acc.requiredPermissions, 0)
			}
		}
	}
	return false
}

// returns true if the write permission of user's global role contains PERMS_CLUSTER_WRITE
func (acc *AccessControl) CanWriteCluster() bool {
	return acc.roles.hasGlobalPermissions(share.PERMS_CLUSTER_READ, share.PERMS_CLUSTER_WRITE)
}

// get all domains over which this access control has the required write permissions
func (acc *AccessControl) GetAdminDomains(writePermitsRequired uint64) []string {
	/*if acc.roles.IsClusterAdmin() || acc.roles.IsFedAdmin() {
		return nil
	} else*/ //->
	if role, ok := acc.roles[AccessDomainGlobal]; ok {
		_, writePermits := getRolePermitValues(role, AccessDomainGlobal)
		if writePermitsRequired == (writePermits & writePermitsRequired) {
			return nil
		}
	}

	list := make([]string, 0)
	for domain, role := range acc.roles {
		_, writePermits := getRolePermitValues(role, domain)
		if writePermitsRequired == (writePermits & writePermitsRequired) {
			list = append(list, domain)
		}
	}
	if len(list) > 0 {
		return list
	} else {
		return nil
	}
}

// Authorize if the access has rights on one of domains which the object is member of.
func (acc *AccessControl) Authorize(obj share.AccessObject, f share.GetAccessObjectFunc) bool {
	var authz bool

	d1, d2 := obj.GetDomain(f) // d1, d2 could contain wildcard character !!
	globalReadPermitsRequired, globalWritePermitsRequired := acc.requiredPermissions, acc.requiredPermissions
	domainReadPermitsRequired, domainWritePermitsRequired := acc.requiredPermissions, acc.requiredPermissions
	if acc.requiredPermissions != share.PERM_IBMSA {
		globalReadPermitsRequired &= share.PERMS_FED_READ
		globalWritePermitsRequired &= share.PERMS_FED_WRITE
		domainReadPermitsRequired &= share.PERMS_DOMAIN_READ
		domainWritePermitsRequired &= share.PERMS_DOMAIN_WRITE
	}

	if d1 == nil && d2 == nil {
		// Global object
		authz = acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
	} else if d2 == nil {
		authz = acc.isAccessAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
	} else {
		if len(d1) == 1 && len(d2) == 1 && d1[0] == share.HiddenFedDomain && d2[0] == share.HiddenFedDomain {
			// d1 & d2 are slice with one entry, share.HiddenFedDomain, if the object requires fed role for access
			authz = acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired|share.PERM_FED)
		} else {
			if acc.op == AccessOPWrite {
				a1 := acc.isAccessAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				a2 := acc.isAccessAllowed(d2, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				authz = a1 && a2
			} else {
				a1 := acc.isAccessAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				a2 := acc.isAccessAllowed(d2, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				authz = a1 || a2
			}
		}
	}

	// if !authz {
	//	  log.WithFields(log.Fields{"resource": reflect.TypeOf(obj), "roles": acc.roles}).Error("Authz failed")
	// }

	return authz
}

// Authorize if the access has rights on all domains which the object is member of.
func (acc *AccessControl) AuthorizeOwn(obj share.AccessObject, f share.GetAccessObjectFunc) bool {
	var authz bool

	d1, d2 := obj.GetDomain(f) // d1, d2 could contain wildcard character !!
	globalReadPermitsRequired, globalWritePermitsRequired := acc.requiredPermissions, acc.requiredPermissions
	domainReadPermitsRequired, domainWritePermitsRequired := acc.requiredPermissions, acc.requiredPermissions
	if acc.requiredPermissions != share.PERM_IBMSA {
		globalReadPermitsRequired &= share.PERMS_FED_READ
		globalWritePermitsRequired &= share.PERMS_FED_WRITE
		domainReadPermitsRequired &= share.PERMS_DOMAIN_READ
		domainWritePermitsRequired &= share.PERMS_DOMAIN_WRITE
	}

	if d1 == nil && d2 == nil {
		// Global object
		authz = acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
	} else if d2 == nil {
		authz = acc.isOwnAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
	} else {
		if len(d1) == 1 && len(d2) == 1 && d1[0] == share.HiddenFedDomain && d2[0] == share.HiddenFedDomain {
			// d1 & d2 are slice with one entry, share.HiddenFedDomain, if the object requires fed role for access
			authz = acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired|share.PERM_FED)
		} else {
			if acc.op == AccessOPWrite {
				a1 := acc.isOwnAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				a2 := acc.isOwnAllowed(d2, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				authz = a1 && a2
			} else {
				a1 := acc.isOwnAllowed(d1, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				a2 := acc.isOwnAllowed(d2, domainReadPermitsRequired, domainWritePermitsRequired) || acc.isOneAccessAllowed(AccessDomainGlobal, globalReadPermitsRequired, globalWritePermitsRequired)
				authz = a1 || a2
			}
		}
	}

	// if !authz {
	// 	log.WithFields(log.Fields{"resource": reflect.TypeOf(obj), "roles": acc.roles}).Error("Authz failed")
	// }

	return authz
}

func (acc *AccessControl) GetRoleDomains() map[string][]string {
	var roleDomains = make(map[string][]string)

	for d, role := range acc.roles {
		roleDomains[role] = append(roleDomains[role], d)
	}

	return roleDomains
}

func ContainsNonSupportRole(role string) bool {
	var roles = utils.NewSet(api.UserRoleFedAdmin, api.UserRoleFedReader, api.UserRoleIBMSA, api.UserRoleImportStatus)
	return roles.Contains(role)
}

func (acc *AccessControl) ExportAccessControl() *api.UserAccessControl {
	c := &api.UserAccessControl{
		Op:                  string(acc.op),
		Roles:               acc.roles,
		WRoles:              acc.wRoles,
		ApiCategoryID:       acc.apiCategoryID,
		RequiredPermissions: acc.requiredPermissions,
		BoostPermissions:    acc.boostPermissions,
	}
	return c
}

func ImportAccessControl(uac *api.UserAccessControl) *AccessControl {
	op := AccessOPRead
	if uac.Op == "write" {
		op = AccessOPWrite
	}
	acc := &AccessControl{
		op:                  op,
		roles:               uac.Roles,
		wRoles:              uac.WRoles,
		apiCategoryID:       uac.ApiCategoryID,
		requiredPermissions: uac.RequiredPermissions,
	}
	return acc
}

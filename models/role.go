package models

import (
	"database/sql"
	"sort"

	"github.com/udovin/gosql"
)

// Role represents a role.
type Role struct {
	baseObject
	// Name contains role name.
	//
	// Name should be unique for all roles in the events.
	Name string `db:"name"`
}

const (
	// LoginRole represents name of role for login action.
	LoginRole = "login"
	// LogoutRole represents name of role for logout action.
	LogoutRole = "logout"
	// RegisterRole represents name of role for register action.
	RegisterRole = "register"
	// StatusRole represents name of role for status check.
	StatusRole = "status"
	// ObserveSettingsRole represents name of role for observing settings.
	ObserveSettingsRole = "observe_settings"
	// CreateSettingRole represents name of role for creating new setting.
	CreateSettingRole = "create_setting"
	// UpdateSettingRole represents name of role for updating setting.
	UpdateSettingRole = "update_setting"
	// DeleteSettingRole represents name of role for deleting setting.
	DeleteSettingRole = "delete_setting"
	// ObserveRolesRole represents name of role for observing roles.
	ObserveRolesRole = "observe_roles"
	// CreateRoleRole represents name of role for creating new role.
	CreateRoleRole = "create_role"
	// DeleteRoleRole represents name of role for deleting role.
	DeleteRoleRole = "delete_role"
	// ObserveRoleRolesRole represents name of role for observing role roles.
	ObserveRoleRolesRole = "observe_role_roles"
	// CreateRoleRoleRole represents name of role for creating new role role.
	CreateRoleRoleRole = "create_role_role"
	// DeleteRoleRoleRole represents name of role for deleting role role.
	DeleteRoleRoleRole = "delete_role_role"
	// ObserveUserRolesRole represents name of role for observing user roles.
	ObserveUserRolesRole = "observe_user_roles"
	// CreateUserRoleRole represents name of role for attaching role to user.
	CreateUserRoleRole = "create_user_role"
	// DeleteUserRoleRole represents name of role for detaching role from user.
	DeleteUserRoleRole = "delete_user_role"
	// ObserveUserRole represents name of role for observing user.
	ObserveUserRole = "observe_user"
	// UpdateUserRole represents name of role for updating user.
	UpdateUserRole = "update_user"
	// ObserveUserEmailRole represents name of role for observing user email.
	ObserveUserEmailRole = "observe_user_email"
	// ObserveUserFirstNameRole represents name of role for observing
	// user first name.
	ObserveUserFirstNameRole = "observe_user_first_name"
	// ObserveUserLastNameRole represents name of role for observing
	// user last name.
	ObserveUserLastNameRole = "observe_user_last_name"
	// ObserveUserMiddleNameRole represents name of role for observing
	// user middle name.
	ObserveUserMiddleNameRole = "observe_user_middle_name"
	// ObserveUserSessionsRole represents name of role for observing
	// user sessions.
	ObserveUserSessionsRole = "observe_user_sessions"
	// UpdateUserPasswordRole represents name of role for updating
	// user password.
	UpdateUserPasswordRole = "update_user_password"
	// UpdateUserEmailRole represents name of role for updating user email.
	UpdateUserEmailRole = "update_user_email"
	// UpdateUserFirstNameRole represents name of role for updating
	// user first name.
	UpdateUserFirstNameRole = "update_user_first_name"
	// UpdateUserLastNameRole represents name of role for updating
	// user last name.
	UpdateUserLastNameRole = "update_user_last_name"
	// UpdateUserMiddleNameRole represents name of role for updating
	// user middle name.
	UpdateUserMiddleNameRole = "update_user_middle_name"
	// ObserveSessionRole represents role for observing session.
	ObserveSessionRole = "observe_session"
	// DeleteSessionRole represents role for deleting session.
	DeleteSessionRole = "delete_session"
	// ObserveProblemsRole represents role for observing problem list.
	ObserveProblemsRole = "observe_problems"
	// ObserveProblemRole represents role for observing problem.
	ObserveProblemRole = "observe_problem"
	// CreateProblemRole represents role for creating problem.
	CreateProblemRole = "create_problem"
	// UpdateProblemRole represents role for updating problem.
	UpdateProblemRole = "update_problem"
	// DeleteProblemRole represents role for deleting problem.
	DeleteProblemRole = "delete_problem"
	// ObserveCompilersRole represents role for observing compiler list.
	ObserveCompilersRole = "observe_compilers"
	// ObserveCompilerRole represents role for observing compiler.
	ObserveCompilerRole = "observe_compiler"
	// CreateCompilerRole represents role for creating compiler.
	CreateCompilerRole = "create_compiler"
	// UpdateCompilerRole represents role for updating compiler.
	UpdateCompilerRole = "update_compiler"
	// DeleteCompilerRole represents role for deleting compiler.
	DeleteCompilerRole = "delete_compiler"
	// ObserveSolutionsRole represents role for observing solution list.
	ObserveSolutionsRole = "observe_solutions"
	// ObserveSolutionRole represents role for observing solution.
	ObserveSolutionRole = "observe_solution"
	//
	ObserveSolutionReportTestNumber = "observe_solution_report_test_number"
	//
	ObserveSolutionReportCheckerLogs = "observe_solution_report_checker_logs"
	// ObserveContestsRole represents role for observing contest list.
	ObserveContestsRole = "observe_contests"
	// ObserveContestRole represents role for observing contest.
	ObserveContestRole = "observe_contest"
	// ObserveContestProblemsRole represents role for observing
	// contest problem list.
	ObserveContestProblemsRole = "observe_contest_problems"
	// ObserveContestProblemRole represents role for observing
	// contest problem.
	ObserveContestProblemRole = "observe_contest_problem"
	// CreateContestProblemRole represents role for creating
	// contest problem.
	CreateContestProblemRole = "create_contest_problem"
	// UpdateContestProblemRole represents role for updating
	// contest problem.
	UpdateContestProblemRole = "update_contest_problem"
	// DeleteContestProblemRole represents role for deleting
	// contest problem.
	DeleteContestProblemRole = "delete_contest_problem"
	// ObserveContestParticipantsRole represents role for observing
	// contest participant list.
	ObserveContestParticipantsRole = "observe_contest_participants"
	// ObserveContestParticipantRole represents role for observing
	// contest participant.
	ObserveContestParticipantRole = "observe_contest_participant"
	// CreateContestProblemRole represents role for creating
	// contest participant.
	CreateContestParticipantRole = "create_contest_participant"
	// DeleteContestParticipantRole represents role for deleting
	// contest participant.
	DeleteContestParticipantRole = "delete_contest_participant"
	// ObserveContestSolutionsRole represents role for observing
	// contest solution list.
	ObserveContestSolutionsRole = "observe_contest_solutions"
	// ObserveContestSolutionRole represents role for observing
	// contest solution.
	ObserveContestSolutionRole = "observe_contest_solution"
	// CreateContestSolutionRole represents role for creating
	// contest solution.
	CreateContestSolutionRole = "create_contest_solution"
	// SubmitContestSolutionRole represents role for submitting
	// contest solution.
	SubmitContestSolutionRole = "submit_contest_solution"
	// UpdateContestSolutionRole represents role for updating
	// contest solution.
	UpdateContestSolutionRole = "update_contest_solution"
	// DeleteContestSolutionRole represents role for deleting
	// contest solution.
	DeleteContestSolutionRole = "delete_contest_solution"
	//
	ObserveContestStandingsRole = "observe_contest_standings"
	//
	ObserveContestFullStandingsRole = "observe_contest_full_standings"
	// CreateContestRole represents role for creating contest.
	CreateContestRole = "create_contest"
	// UpdateContestRole represents role for updating contest.
	UpdateContestRole = "update_contest"
	// DeleteContestRole represents role for deleting contest.
	DeleteContestRole = "delete_contest"
	// RegisterContestsRole represents role for register to contests.
	RegisterContestsRole = "register_contests"
	// RegisterContestRole represents role for register to contest.
	RegisterContestRole = "register_contest"
	// DeregisterContestRole represents role for deregister from contest.
	DeregisterContestRole = "deregister_contest"
	// ObserveFileContentRole represents role for observing file content.
	ObserveFileContentRole = "observe_file_content"
	//
	ObserveScopesRole            = "observe_scopes"
	ObserveScopeRole             = "observe_scope"
	CreateScopeRole              = "create_scope"
	UpdateScopeRole              = "update_scope"
	DeleteScopeRole              = "delete_scope"
	ObserveScopeUserRole         = "observe_scope_user"
	ObserveScopeUserPasswordRole = "observe_scope_user_password"
	CreateScopeUserRole          = "create_scope_user"
	UpdateScopeUserRole          = "update_scope_user"
	DeleteScopeUserRole          = "delete_scope_user"
)

var builtInRoles = map[string]struct{}{
	LoginRole:                        {},
	LogoutRole:                       {},
	RegisterRole:                     {},
	StatusRole:                       {},
	ObserveSettingsRole:              {},
	CreateSettingRole:                {},
	UpdateSettingRole:                {},
	DeleteSettingRole:                {},
	ObserveRolesRole:                 {},
	CreateRoleRole:                   {},
	DeleteRoleRole:                   {},
	ObserveRoleRolesRole:             {},
	CreateRoleRoleRole:               {},
	DeleteRoleRoleRole:               {},
	ObserveUserRolesRole:             {},
	CreateUserRoleRole:               {},
	DeleteUserRoleRole:               {},
	ObserveUserRole:                  {},
	UpdateUserRole:                   {},
	ObserveUserEmailRole:             {},
	ObserveUserFirstNameRole:         {},
	ObserveUserLastNameRole:          {},
	ObserveUserMiddleNameRole:        {},
	ObserveUserSessionsRole:          {},
	UpdateUserPasswordRole:           {},
	UpdateUserEmailRole:              {},
	UpdateUserFirstNameRole:          {},
	UpdateUserLastNameRole:           {},
	UpdateUserMiddleNameRole:         {},
	ObserveSessionRole:               {},
	ObserveProblemsRole:              {},
	ObserveProblemRole:               {},
	CreateProblemRole:                {},
	UpdateProblemRole:                {},
	DeleteProblemRole:                {},
	ObserveCompilersRole:             {},
	ObserveCompilerRole:              {},
	CreateCompilerRole:               {},
	UpdateCompilerRole:               {},
	DeleteCompilerRole:               {},
	ObserveSolutionsRole:             {},
	ObserveSolutionRole:              {},
	ObserveSolutionReportTestNumber:  {},
	ObserveSolutionReportCheckerLogs: {},
	ObserveContestRole:               {},
	ObserveContestProblemsRole:       {},
	ObserveContestProblemRole:        {},
	CreateContestProblemRole:         {},
	UpdateContestProblemRole:         {},
	DeleteContestProblemRole:         {},
	ObserveContestParticipantsRole:   {},
	ObserveContestParticipantRole:    {},
	CreateContestParticipantRole:     {},
	DeleteContestParticipantRole:     {},
	ObserveContestSolutionsRole:      {},
	ObserveContestSolutionRole:       {},
	CreateContestSolutionRole:        {},
	SubmitContestSolutionRole:        {},
	UpdateContestSolutionRole:        {},
	DeleteContestSolutionRole:        {},
	ObserveContestStandingsRole:      {},
	ObserveContestFullStandingsRole:  {},
	ObserveContestsRole:              {},
	CreateContestRole:                {},
	UpdateContestRole:                {},
	DeleteContestRole:                {},
	DeleteSessionRole:                {},
	RegisterContestsRole:             {},
	RegisterContestRole:              {},
	DeregisterContestRole:            {},
	ObserveFileContentRole:           {},
	ObserveScopesRole:                {},
	ObserveScopeRole:                 {},
	CreateScopeRole:                  {},
	UpdateScopeRole:                  {},
	DeleteScopeRole:                  {},
	ObserveScopeUserRole:             {},
	ObserveScopeUserPasswordRole:     {},
	CreateScopeUserRole:              {},
	UpdateScopeUserRole:              {},
	DeleteScopeUserRole:              {},
}

// GetBuildInRoles returns all built-in roles.
func GetBuiltInRoles() []string {
	var roles []string
	for role := range builtInRoles {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

// IsBuiltIn returns flag that role is built-in.
func (o Role) IsBuiltIn() bool {
	_, ok := builtInRoles[o.Name]
	return ok
}

// Clone creates copy of role.
func (o Role) Clone() Role {
	return o
}

// RoleEvent represents role event.
type RoleEvent struct {
	baseEvent
	Role
}

// Object returns event role.
func (e RoleEvent) Object() Role {
	return e.Role
}

// SetObject sets event role.
func (e *RoleEvent) SetObject(o Role) {
	e.Role = o
}

// RoleStore represents a role store.
type RoleStore struct {
	baseStore[Role, RoleEvent, *Role, *RoleEvent]
	byName *index[string, Role, *Role]
}

// GetByName returns role by name.
//
// If there is no role with specified name then
// sql.ErrNoRows will be returned.
func (s *RoleStore) GetByName(name string) (Role, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for id := range s.byName.Get(name) {
		if object, ok := s.objects[id]; ok {
			return object.Clone(), nil
		}
	}
	return Role{}, sql.ErrNoRows
}

var _ baseStoreImpl[Role] = (*RoleStore)(nil)

// NewRoleStore creates a new instance of RoleStore.
func NewRoleStore(
	db *gosql.DB, table, eventTable string,
) *RoleStore {
	impl := &RoleStore{
		byName: newIndex(func(o Role) string { return o.Name }),
	}
	impl.baseStore = makeBaseStore[Role, RoleEvent](
		db, table, eventTable, impl, impl.byName,
	)
	return impl
}

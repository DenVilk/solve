package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Role struct {
	ID   int64  `db:"id"   json:""`
	Code string `db:"code" json:""`
}

type RoleChange struct {
	BaseChange
	Role
}

type RoleStore struct {
	Manager     *ChangeManager
	db          *sql.DB
	table       string
	changeTable string
	roles       map[int64]Role
}

func (c *RoleChange) ChangeData() interface{} {
	return c.Role
}

func NewRoleStore(
	db *sql.DB, table, changeTable string,
) *RoleStore {
	store := RoleStore{
		db: db, table: table, changeTable: changeTable,
		roles: make(map[int64]Role),
	}
	store.Manager = NewChangeManager(&store)
	return &store
}

func (s *RoleStore) GetDB() *sql.DB {
	return s.db
}

func (s *RoleStore) ChangeTableName() string {
	return s.changeTable
}

func (s *RoleStore) Get(id int64) (Role, bool) {
	role, ok := s.roles[id]
	return role, ok
}

func (s *RoleStore) Create(m *Role) error {
	change := RoleChange{
		BaseChange: BaseChange{Type: CreateChange},
		Role:       *m,
	}
	err := s.Manager.Change(&change)
	if err != nil {
		return err
	}
	*m = change.Role
	return nil
}

func (s *RoleStore) Update(m *Role) error {
	change := RoleChange{
		BaseChange: BaseChange{Type: UpdateChange},
		Role:       *m,
	}
	err := s.Manager.Change(&change)
	if err != nil {
		return err
	}
	*m = change.Role
	return nil
}

func (s *RoleStore) Delete(id int64) error {
	change := RoleChange{
		BaseChange: BaseChange{Type: DeleteChange},
		Role:       Role{ID: id},
	}
	return s.Manager.Change(&change)
}

func (s *RoleStore) loadChangeGapTx(
	tx *ChangeTx, gap ChangeGap,
) (*sql.Rows, error) {
	return tx.Query(
		fmt.Sprintf(
			`SELECT`+
				` "change_id", "change_type", "change_time",`+
				` "id", "code"`+
				` FROM "%s"`+
				` WHERE "change_id" >= $1 AND "change_id" < $2`+
				` ORDER BY "change_id"`,
			s.ChangeTableName(),
		),
		gap.BeginID, gap.EndID,
	)
}

func (s *RoleStore) scanChange(scan Scanner) (Change, error) {
	change := &RoleChange{}
	err := scan.Scan(
		&change.BaseChange.ID, &change.Type, &change.Time,
		&change.Role.ID, &change.Code,
	)
	return change, err
}

func (s *RoleStore) saveChangeTx(tx *ChangeTx, change Change) error {
	role := change.(*RoleChange)
	role.Time = time.Now().Unix()
	switch role.Type {
	case CreateChange:
		res, err := tx.Exec(
			fmt.Sprintf(
				`INSERT INTO "%s" ("code") VALUES ($1)`,
				s.table,
			),
			role.Code,
		)
		if err != nil {
			return err
		}
		role.Role.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	case UpdateChange:
		if _, ok := s.roles[role.Role.ID]; !ok {
			return fmt.Errorf(
				"role with id = %d does not exists",
				role.Role.ID,
			)
		}
		_, err := tx.Exec(
			fmt.Sprintf(
				`UPDATE "%s" SET "code" = $1 WHERE "id" = $2`,
				s.table,
			),
			role.Code, role.Role.ID,
		)
		if err != nil {
			return err
		}
	case DeleteChange:
		if _, ok := s.roles[role.Role.ID]; !ok {
			return fmt.Errorf(
				"role with id = %d does not exists",
				role.Role.ID,
			)
		}
		_, err := tx.Exec(
			fmt.Sprintf(
				`DELETE FROM "%s" WHERE "id" = $1`,
				s.table,
			),
			role.Role.ID,
		)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"unsupported change type = %s",
			role.Type,
		)
	}
	res, err := tx.Exec(
		fmt.Sprintf(
			`INSERT INTO "%s"`+
				` ("change_type", "change_time", "id", "code")`+
				` VALUES ($1, $2, $3, $4)`,
			s.ChangeTableName(),
		),
		role.Type, role.Time, role.Role.ID, role.Code,
	)
	if err != nil {
		return err
	}
	role.BaseChange.ID, err = res.LastInsertId()
	return err
}

func (s *RoleStore) applyChange(change Change) {
	roleChange := change.(*RoleChange)
	role := roleChange.Role
	switch roleChange.Type {
	case UpdateChange:
		fallthrough
	case CreateChange:
		s.roles[role.ID] = role
	case DeleteChange:
		delete(s.roles, role.ID)
	default:
		panic(fmt.Errorf(
			"unsupported change type = %s",
			roleChange.Type,
		))
	}
}
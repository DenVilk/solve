package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"

	"golang.org/x/crypto/sha3"

	"github.com/udovin/solve/db"
)

// User contains common information about user.
type User struct {
	ID           int64  `db:"id"`
	AccountID    int64  `db:"account_id"`
	Login        string `db:"login"`
	PasswordHash string `db:"password_hash"`
	PasswordSalt string `db:"password_salt"`
}

// ObjectID returns ID of user.
func (o User) ObjectID() int64 {
	return o.ID
}

func (o User) clone() User {
	return o
}

// UserEvent represents an user event.
type UserEvent struct {
	baseEvent
	User
}

// Object returns user.
func (e UserEvent) Object() db.Object {
	return e.User
}

// WithObject return copy of event with replaced user.
func (e UserEvent) WithObject(o db.Object) ObjectEvent {
	e.User = o.(User)
	return e
}

// UserStore represents users store.
type UserStore struct {
	baseStore
	users     map[int64]User
	byAccount map[int64]int64
	byLogin   map[string]int64
	salt      string
}

// Get returns user by ID.
func (s *UserStore) Get(id int64) (User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if user, ok := s.users[id]; ok {
		return user.clone(), nil
	}
	return User{}, sql.ErrNoRows
}

// GetByLogin returns user by login.
func (s *UserStore) GetByLogin(login string) (User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if id, ok := s.byLogin[login]; ok {
		if user, ok := s.users[id]; ok {
			return user.clone(), nil
		}
	}
	return User{}, sql.ErrNoRows
}

// GetByAccount returns user by login.
func (s *UserStore) GetByAccount(id int64) (User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if id, ok := s.byAccount[id]; ok {
		if user, ok := s.users[id]; ok {
			return user.clone(), nil
		}
	}
	return User{}, sql.ErrNoRows
}

// CreateTx creates user and returns copy with valid ID.
func (s *UserStore) CreateTx(tx *sql.Tx, user User) (User, error) {
	event, err := s.createObjectEvent(tx, UserEvent{
		makeBaseEvent(CreateEvent),
		user,
	})
	if err != nil {
		return User{}, err
	}
	return event.Object().(User), nil
}

// UpdateTx updates user with specified ID.
func (s *UserStore) UpdateTx(tx *sql.Tx, user User) error {
	_, err := s.createObjectEvent(tx, UserEvent{
		makeBaseEvent(UpdateEvent),
		user,
	})
	return err
}

// DeleteTx deletes user with specified ID.
func (s *UserStore) DeleteTx(tx *sql.Tx, id int64) error {
	_, err := s.createObjectEvent(tx, UserEvent{
		makeBaseEvent(DeleteEvent),
		User{ID: id},
	})
	return err
}

// SetPassword modifies PasswordHash and PasswordSalt fields.
//
// PasswordSalt will be replaced with random 16 byte string
// and PasswordHash will be calculated using password, salt
// and global salt.
func (s *UserStore) SetPassword(user *User, password string) error {
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return err
	}
	user.PasswordSalt = encodeBase64(saltBytes)
	user.PasswordHash = hashPassword(password, user.PasswordSalt, s.salt)
	return nil
}

// CheckPassword checks that passwords are the same.
func (s *UserStore) CheckPassword(user User, password string) bool {
	passwordHash := hashPassword(password, user.PasswordSalt, s.salt)
	return passwordHash == user.PasswordHash
}

func (s *UserStore) reset() {
	s.users = map[int64]User{}
	s.byAccount = map[int64]int64{}
	s.byLogin = map[string]int64{}
}

func (s *UserStore) onCreateObject(o db.Object) {
	user := o.(User)
	s.users[user.ID] = user
	s.byAccount[user.AccountID] = user.ID
	s.byLogin[user.Login] = user.ID
}

func (s *UserStore) onDeleteObject(o db.Object) {
	user := o.(User)
	delete(s.byAccount, user.AccountID)
	delete(s.byLogin, user.Login)
	delete(s.users, user.ID)
}

func (s *UserStore) onUpdateObject(o db.Object) {
	user := o.(User)
	if old, ok := s.users[user.ID]; ok {
		if old.AccountID != user.AccountID {
			delete(s.byAccount, old.AccountID)
		}
		if old.Login != user.Login {
			delete(s.byLogin, old.Login)
		}
	}
	s.onCreateObject(o)
}

// NewUserStore creates new instance of user store.
func NewUserStore(
	table, eventTable, salt string, dialect db.Dialect,
) *UserStore {
	impl := &UserStore{salt: salt}
	impl.baseStore = makeBaseStore(
		User{}, table, UserEvent{}, eventTable, impl, dialect,
	)
	return impl
}

func hashPassword(password, salt, globalSalt string) string {
	return hashString(salt + hashString(password) + globalSalt)
}

func encodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func hashString(value string) string {
	bytes := sha3.Sum512([]byte(value))
	return encodeBase64(bytes[:])
}

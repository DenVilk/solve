package models

import (
	"database/sql"
	"testing"

	"github.com/udovin/gosql"
	"github.com/udovin/solve/db"
)

type userStoreTest struct{}

func (t *userStoreTest) prepareDB(tx *sql.Tx) error {
	if _, err := tx.Exec(
		`CREATE TABLE "user" (` +
			`"id" integer PRIMARY KEY,` +
			`"account_id" integer NOT NULL,` +
			`"login" varchar(64) NOT NULL,` +
			`"password_hash" varchar(255) NOT NULL,` +
			`"password_salt" varchar(255) NOT NULL,` +
			`"email" varchar(255),` +
			`"first_name" varchar(255),` +
			`"last_name" varchar(255),` +
			`"middle_name" varchar(255))`,
	); err != nil {
		return err
	}
	_, err := tx.Exec(
		`CREATE TABLE "user_event" (` +
			`"event_id" integer PRIMARY KEY,` +
			`"event_type" int8 NOT NULL,` +
			`"event_time" bigint NOT NULL,` +
			`"id" integer NOT NULL,` +
			`"account_id" integer NOT NULL,` +
			`"login" varchar(64) NOT NULL,` +
			`"password_hash" varchar(255) NOT NULL,` +
			`"password_salt" varchar(255) NOT NULL,` +
			`"email" varchar(255),` +
			`"first_name" varchar(255),` +
			`"last_name" varchar(255),` +
			`"middle_name" varchar(255))`,
	)
	return err
}

func (t *userStoreTest) newStore() Store {
	return NewUserStore("user", "user_event", "", gosql.SQLiteDialect)
}

func (t *userStoreTest) newObject() db.Object {
	return User{}
}

func (t *userStoreTest) createObject(
	s Store, tx *sql.Tx, o db.Object,
) (db.Object, error) {
	user := o.(User)
	if err := s.(*UserStore).CreateTx(tx, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (t *userStoreTest) updateObject(
	s Store, tx *sql.Tx, o db.Object,
) (db.Object, error) {
	return o, s.(*UserStore).UpdateTx(tx, o.(User))
}

func (t *userStoreTest) deleteObject(
	s Store, tx *sql.Tx, id int64,
) error {
	return s.(*UserStore).DeleteTx(tx, id)
}

func TestUserStore(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	tester := StoreTester{&userStoreTest{}}
	tester.Test(t)
}

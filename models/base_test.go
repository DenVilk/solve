package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/udovin/solve/db"
)

type testObjectBase struct {
	String string `db:"string"`
	Int    int    `db:"int"`
	UInt   uint   `db:"uint"`
	Bool   bool   `db:"bool"`
	Bytes  []byte `db:"bytes"`
}

type testJSON struct {
	Text string `json:""`
}

func (s testJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *testJSON) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	case nil:
		*s = testJSON{}
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

type testObject struct {
	testObjectBase
	ID   int64    `db:"id"`
	JSON testJSON `db:"json"`
}

func (o testObject) ObjectID() int64 {
	return o.ID
}

type testObjectEvent struct {
	baseEvent
	testObject
}

func (e testObjectEvent) Object() db.Object {
	return e.testObject
}

func (e testObjectEvent) WithObject(o db.Object) ObjectEvent {
	e.testObject = o.(testObject)
	return e
}

type testManager struct {
	baseManager
	table, eventTable string
	objects           map[int64]testObject
}

func (m *testManager) Get(id int64) (testObject, error) {
	if object, ok := m.objects[id]; ok {
		return object, nil
	}
	return testObject{}, sql.ErrNoRows
}

func (m *testManager) CreateTx(
	tx *sql.Tx, object testObject,
) (testObject, error) {
	event, err := m.createObjectEvent(tx, testObjectEvent{
		makeBaseEvent(CreateEvent),
		object,
	})
	if err != nil {
		return testObject{}, err
	}
	return event.Object().(testObject), nil
}

func (m *testManager) UpdateTx(tx *sql.Tx, object testObject) error {
	_, err := m.createObjectEvent(tx, testObjectEvent{
		makeBaseEvent(UpdateEvent),
		object,
	})
	return err
}

func (m *testManager) DeleteTx(tx *sql.Tx, id int64) error {
	_, err := m.createObjectEvent(tx, testObjectEvent{
		makeBaseEvent(DeleteEvent),
		testObject{ID: id},
	})
	return err
}

func (m *testManager) reset() {
	m.objects = map[int64]testObject{}
}

func (m *testManager) addObject(o db.Object) {
	m.objects[o.ObjectID()] = o.(testObject)
}

func (m *testManager) onCreateObject(o db.Object) {
	if _, ok := m.objects[o.ObjectID()]; ok {
		panic("object already exists")
	}
	m.objects[o.ObjectID()] = o.(testObject)
}

func (m *testManager) onUpdateObject(o db.Object) {
	if _, ok := m.objects[o.ObjectID()]; !ok {
		panic("object not found")
	}
	m.objects[o.ObjectID()] = o.(testObject)
}

func (m *testManager) onDeleteObject(o db.Object) {
	if _, ok := m.objects[o.ObjectID()]; !ok {
		panic("object not found")
	}
	delete(m.objects, o.ObjectID())
}

func migrateTestManager(t testing.TB, m *testManager) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal("Error:", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.Exec(fmt.Sprintf(
		`CREATE TABLE %q (`+
			`"id" integer PRIMARY KEY,`+
			`"string" varchar(255) NOT NULL,`+
			`"int" integer NOT NULL,`+
			`"uint" integer NOT NULL,`+
			`"bool" boolean NOT NULL,`+
			`"bytes" blob,`+
			`"json" blob NOT NULL)`,
		m.table,
	)); err != nil {
		t.Fatal("Error:", err)
	}
	if _, err := tx.Exec(fmt.Sprintf(
		`CREATE TABLE %q (`+
			`"event_id" integer PRIMARY KEY,`+
			`"event_type" int8 NOT NULL,`+
			`"event_time" bigint NOT NULL,`+
			`"id" integer NOT NULL,`+
			`"string" varchar(255) NOT NULL,`+
			`"int" integer NOT NULL,`+
			`"uint" integer NOT NULL,`+
			`"bool" boolean NOT NULL,`+
			`"bytes" blob,`+
			`"json" blob NOT NULL)`,
		m.eventTable,
	)); err != nil {
		t.Fatal("Error:", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal("Error:", err)
	}
}

func newTestManager() *testManager {
	impl := &testManager{
		table:      "test_object",
		eventTable: "test_object_event",
	}
	impl.baseManager = makeBaseManager(
		testObject{}, impl.table,
		testObjectEvent{}, impl.eventTable,
		impl, db.SQLite,
	)
	return impl
}

func testInitManager(t testing.TB, m Manager) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := m.InitTx(tx); err != nil {
		t.Fatal(err)
	}
}

func testSyncManager(t testing.TB, m Manager) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := m.SyncTx(tx); err != nil {
		t.Fatal(err)
	}
}

func createTestObject(t testing.TB, m *testManager, o testObject) testObject {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal("Error:", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if o, err = m.CreateTx(tx, o); err != nil {
		t.Fatal("Error:", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal("Error:", err)
	}
	return o
}

func updateTestObject(
	t testing.TB, m *testManager, o testObject, expErr error,
) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = m.UpdateTx(tx, o); err != expErr {
		t.Fatalf("Expected %v, got %v", expErr, err)
	}
	if err == nil {
		if err := tx.Commit(); err != nil {
			t.Fatal("Error:", err)
		}
	}
}

func deleteTestObject(
	t testing.TB, m *testManager, id int64, expErr error,
) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	err = m.DeleteTx(tx, id)
	if err != expErr {
		t.Fatal(err)
	}
	if err == nil {
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMakeBaseManager(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	master := newTestManager()
	replica := newTestManager()
	migrateTestManager(t, master)
	testInitManager(t, master)
	testInitManager(t, replica)
	object := testObject{
		testObjectBase: testObjectBase{
			String: "Test",
			Int:    822,
			UInt:   232,
			Bool:   true,
			Bytes:  []byte{8, 1, 4, 8},
		},
		JSON: testJSON{Text: "Test message"},
	}
	savedObject := createTestObject(t, master, object)
	if object.ID == savedObject.ID {
		t.Fatalf("IDs should be different: %v", object.ID)
	}
	if _, err := replica.Get(savedObject.ID); err != sql.ErrNoRows {
		t.Fatalf(
			"Replica already contains object: %v", savedObject.ID,
		)
	}
	checkReplicaObject := func(object testObject, expErr error) {
		testSyncManager(t, replica)
		loaded, err := replica.Get(object.ID)
		if err != expErr {
			t.Fatalf(
				"Replica does not contain object: %v", object.ID,
			)
		}
		if err == nil {
			if !reflect.DeepEqual(loaded, object) {
				t.Fatalf(
					"Objects are not equal: %v != %v", loaded, object,
				)
			}
		}
	}
	checkReplicaObject(savedObject, nil)
	savedObject.Int = 12345
	savedObject.JSON = testJSON{Text: "Updated message"}
	updateTestObject(t, master, savedObject, nil)
	checkReplicaObject(savedObject, nil)
	updateTestObject(t, master, testObject{ID: 100}, sql.ErrNoRows)
	deleteTestObject(t, master, savedObject.ID, nil)
	deleteTestObject(t, master, savedObject.ID, sql.ErrNoRows)
	checkReplicaObject(savedObject, sql.ErrNoRows)
}

func TestBaseManager_lockStore(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	manager := newTestManager()
	migrateTestManager(t, manager)
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal("Error:", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := manager.lockStore(tx); err != nil {
		t.Fatal("Error:", err)
	}
}

func TestBaseEvent(t *testing.T) {
	ts := time.Now()
	event := baseEvent{BaseEventTime: ts.Unix()}
	if v := event.EventTime(); ts.Sub(v) > time.Second {
		t.Fatalf("Expected %v, got %v", ts, v)
	}
}

func TestNInt64_Value(t *testing.T) {
	var a NInt64 = 0
	va, err := a.Value()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if va != nil {
		t.Fatalf("Expected %v, got %v", nil, va)
	}
	var b NInt64 = 12345
	vb, err := b.Value()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if vb.(int64) != 12345 {
		t.Fatalf("Expected %v, got %v", 12345, vb)
	}
}

func TestNInt64_Scan(t *testing.T) {
	var a NInt64 = 12345
	if err := a.Scan(nil); err != nil {
		t.Fatal("Error:", err)
	}
	if a != 0 {
		t.Fatalf("Expected %v, got %v", 0, a)
	}
	var b NInt64 = 0
	if err := b.Scan(int64(12345)); err != nil {
		t.Fatal("Error:", err)
	}
	if b != 12345 {
		t.Fatalf("Expected %v, got %v", 12345, b)
	}
	var c NInt64 = 0
	if err := c.Scan(false); err == nil {
		t.Fatal("Expected error")
	}
}
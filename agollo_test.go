package agollo

import (
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ZhengHe-MD/agollo/internal/mockserver"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	go func() {
		if err := mockserver.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	// wait for mock server to run
	time.Sleep(time.Millisecond * 10)
}

func teardown() {
	mockserver.Close()
}

func TestAgolloStart(t *testing.T) {
	if err := Start(); err == nil {
		t.Errorf("Start with default app.properties should return err, got :%v", err)
		return
	}

	if err := StartWithConfFile("fake.properties"); err == nil {
		t.Errorf("Start with fake.properties should return err, got :%v", err)
		return
	}

	if err := StartWithConfFile("./testdata/app.properties"); err != nil {
		t.Errorf("Start with app.properties should return nil, got :%v", err)
		return
	}

	f, err := os.Stat(path.Dir(defaultClient.getDumpFileName()))
	if err != nil {
		t.Errorf("dump file dir should exists, got err:%v", err)
		return
	}

	if !f.IsDir() {
		t.Errorf("dump file dir should be a dir, got file")
		return
	}

	if err := Stop(); err != nil {
		t.Errorf("Stop should return nil, got :%v", err)
		return
	}
	os.Remove(defaultClient.getDumpFileName())

	if err := StartWithConfFile("./testdata/app.properties"); err != nil {
		t.Errorf("Start with app.properties should return nil, got :%v", err)
		return
	}
	defer Stop()
	defer os.Remove(defaultClient.getDumpFileName())

	if err := defaultClient.loadLocal(defaultClient.getDumpFileName()); err != nil {
		t.Errorf("loadLocal should return nil, got: %v", err)
		return
	}

	mockserver.Set("application", "key", "value")
	updates := WatchUpdate()

	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val := GetString("key", "defaultValue")
	if val != "value" {
		t.Errorf("GetString of key should = value, got %v", val)
		return
	}

	keys := GetAllKeys("application")
	if len(keys) != 1 {
		t.Errorf("GetAllKeys should return 1 key")
		return
	}

	mockserver.Set("application", "key", "newvalue")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = defaultClient.GetString("key", "defaultValue")
	if val != "newvalue" {
		t.Errorf("GetString of key should = newvalue, got %v", val)
		return
	}

	keys = GetAllKeys("application")
	if len(keys) != 1 {
		t.Errorf("GetAllKeys should return 1 key")
		return
	}

	mockserver.Delete("application", "key")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = GetString("key", "defaultValue")
	if val != "defaultValue" {
		t.Errorf("GetString of key should = defaultValue, got %v", val)
		return
	}

	keys = GetAllKeys("application")
	if len(keys) != 0 {
		t.Errorf("GetAllKeys should return 0 key")
		return
	}

	mockserver.Set("client.json", "content", `{"name":"agollo"}`)
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = GetNameSpaceContent("client.json", "{}")
	if val != `{"name":"agollo"}` {
		t.Errorf(`GetString of client.json content should  = {"name":"agollo"}, got %v`, val)
		return
	}

	if err := SubscribeToNamespaces("new_namespace.json"); err != nil {
		t.Error(err)
		return
	}

	mockserver.Set("new_namespace.json", "key", "1")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = GetStringWithNamespace("new_namespace.json", "key", "defaultValue")
	if val != `1` {
		t.Errorf(`GetStringWithNamespace of new_namespace.json content should  = 1, got %v`, val)
		return
	}


	intVal := GetIntWithNamespace("new_namespace.json", "key", 0)
	if intVal != 1 {
		t.Errorf(`GetIntWithNamespace of new_namespace.json content should = 1, got %v`, intVal)
		return
	}

	boolVal := GetBoolWithNamespace("new_namespace.json", "key", false)
	if !boolVal {
		t.Errorf(`GetBoolWithNamespace of new_namespace.json content should = true, got %v`, boolVal)
		return
	}

	float64Val := GetFloat64WithNamespace("new_namespace.json", "key", float64(0))
	if float64Val != float64(1) {
		t.Errorf(`GetFloat64WithNamespace of new_namespace.json content should = 1, got %v`, float64Val)
		return
	}
}

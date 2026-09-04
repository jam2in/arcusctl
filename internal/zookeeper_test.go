package internal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-zookeeper/zk"
)

func TestParseZooKeeperAddresses(t *testing.T) {
	testcases := []struct {
		name      string
		addresses string
		want      []string
		wantErr   bool
	}{
		{name: "single", addresses: "localhost:2181", want: []string{"localhost:2181"}},
		{
			name:      "multiple with whitespace",
			addresses: "zk1:2181, zk2:2181 ,zk3:2181",
			want:      []string{"zk1:2181", "zk2:2181", "zk3:2181"},
		},
		{name: "empty", addresses: "", wantErr: true},
		{name: "empty member", addresses: "zk1:2181,,zk2:2181", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseZooKeeperAddresses(tc.addresses)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseZooKeeperAddresses(%q) error = %v, wantErr %v", tc.addresses, err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseZooKeeperAddresses(%q) = %v, want %v", tc.addresses, got, tc.want)
			}
		})
	}
}

func TestEnsureZNodeCreatesMissingParents(t *testing.T) {
	client := newFakeZNodeClient()
	client.existing["/arcus"] = true

	if err := ensureZNode(client, "/arcus/cache/list"); err != nil {
		t.Fatalf("ensureZNode() error = %v", err)
	}

	wantCreated := []string{"/arcus/cache", "/arcus/cache/list"}
	if !reflect.DeepEqual(client.created, wantCreated) {
		t.Errorf("created nodes = %v, want %v", client.created, wantCreated)
	}
}

func TestEnsureZNodeHandlesConcurrentCreate(t *testing.T) {
	client := newFakeZNodeClient()
	client.createErrors["/arcus"] = zk.ErrNodeExists

	if err := ensureZNode(client, "/arcus/cache"); err != nil {
		t.Fatalf("ensureZNode() error = %v", err)
	}

	wantCreated := []string{"/arcus", "/arcus/cache"}
	if !reflect.DeepEqual(client.created, wantCreated) {
		t.Errorf("create attempts = %v, want %v", client.created, wantCreated)
	}
}

func TestEnsureZNodeWrapsOperationError(t *testing.T) {
	wantErr := errors.New("connection lost")
	client := newFakeZNodeClient()
	client.existsErrors["/arcus/cache"] = wantErr

	err := ensureZNode(client, "/arcus/cache")
	if !errors.Is(err, wantErr) {
		t.Fatalf("ensureZNode() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestDeleteZNodeDeletesChildrenBeforeParent(t *testing.T) {
	client := newFakeZNodeClient()
	client.children["/arcus"] = []string{"cache", "clients"}
	client.children["/arcus/cache"] = []string{"service"}

	if err := deleteZNode(client, "/arcus"); err != nil {
		t.Fatalf("deleteZNode() error = %v", err)
	}

	wantDeleted := []string{
		"/arcus/cache/service",
		"/arcus/cache",
		"/arcus/clients",
		"/arcus",
	}
	if !reflect.DeepEqual(client.deleted, wantDeleted) {
		t.Errorf("deleted nodes = %v, want %v", client.deleted, wantDeleted)
	}
}

func TestDeleteZNodeTreatsMissingNodeAsSuccess(t *testing.T) {
	client := newFakeZNodeClient()
	client.childrenErrors["/missing"] = zk.ErrNoNode

	if err := deleteZNode(client, "/missing"); err != nil {
		t.Fatalf("deleteZNode() error = %v, want nil", err)
	}
	if len(client.deleted) != 0 {
		t.Errorf("deleted nodes = %v, want none", client.deleted)
	}
}

func TestValidateZNodePath(t *testing.T) {
	testcases := []struct {
		name    string
		zpath   string
		wantErr bool
	}{
		{name: "root", zpath: "/"},
		{name: "node", zpath: "/arcus/cache"},
		{name: "empty", zpath: "", wantErr: true},
		{name: "relative", zpath: "arcus/cache", wantErr: true},
		{name: "trailing slash", zpath: "/arcus/", wantErr: true},
		{name: "duplicate slash", zpath: "/arcus//cache", wantErr: true},
		{name: "dot segment", zpath: "/arcus/../cache", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateZNodePath(tc.zpath)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateZNodePath(%q) error = %v, wantErr %v", tc.zpath, err, tc.wantErr)
			}
		})
	}
}

func TestZNodePublicFunctionsRejectUnsafeArguments(t *testing.T) {
	if err := EnsureZNode(nil, "/arcus"); err == nil {
		t.Error("EnsureZNode(nil, ...) error = nil, want error")
	}
	if err := DeleteZNode(nil, "/arcus"); err == nil {
		t.Error("DeleteZNode(nil, ...) error = nil, want error")
	}
	if err := DeleteZNode(&zk.Conn{}, "/"); err == nil {
		t.Error("DeleteZNode(..., root) error = nil, want error")
	}
}

type fakeZNodeClient struct {
	existing       map[string]bool
	children       map[string][]string
	existsErrors   map[string]error
	createErrors   map[string]error
	childrenErrors map[string]error
	deleteErrors   map[string]error
	created        []string
	deleted        []string
}

func newFakeZNodeClient() *fakeZNodeClient {
	return &fakeZNodeClient{
		existing:       make(map[string]bool),
		children:       make(map[string][]string),
		existsErrors:   make(map[string]error),
		createErrors:   make(map[string]error),
		childrenErrors: make(map[string]error),
		deleteErrors:   make(map[string]error),
	}
}

func (f *fakeZNodeClient) Exists(zpath string) (bool, *zk.Stat, error) {
	if err := f.existsErrors[zpath]; err != nil {
		return false, nil, err
	}
	return f.existing[zpath], nil, nil
}

func (f *fakeZNodeClient) Create(
	zpath string,
	_ []byte,
	_ int32,
	_ []zk.ACL,
) (string, error) {
	f.created = append(f.created, zpath)
	if err := f.createErrors[zpath]; err != nil {
		if errors.Is(err, zk.ErrNodeExists) {
			f.existing[zpath] = true
		}
		return "", err
	}
	f.existing[zpath] = true
	return zpath, nil
}

func (f *fakeZNodeClient) Children(zpath string) ([]string, *zk.Stat, error) {
	if err := f.childrenErrors[zpath]; err != nil {
		return nil, nil, err
	}
	return f.children[zpath], nil, nil
}

func (f *fakeZNodeClient) Delete(zpath string, _ int32) error {
	if err := f.deleteErrors[zpath]; err != nil {
		return err
	}
	f.deleted = append(f.deleted, zpath)
	return nil
}

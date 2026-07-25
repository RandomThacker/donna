package business_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
)

type mockHealthRepo struct {
	err error
}

func (m mockHealthRepo) Ping(context.Context) error {
	return m.err
}

func TestLivenessAlwaysOK(t *testing.T) {
	svc := business.NewHealthService(mockHealthRepo{}, business.BuildInfo{})
	got := svc.Liveness()
	if got.Status != constant.StatusOK {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestReadinessOK(t *testing.T) {
	svc := business.NewHealthService(mockHealthRepo{}, business.BuildInfo{})
	status, err := svc.Readiness(context.Background())
	if err != nil {
		t.Fatalf("Readiness: %v", err)
	}
	if status.Status != constant.StatusOK || status.Database != constant.DatabaseUp {
		t.Fatalf("status = %#v", status)
	}
}

func TestReadinessUnavailable(t *testing.T) {
	svc := business.NewHealthService(mockHealthRepo{err: errors.New("down")}, business.BuildInfo{})
	status, err := svc.Readiness(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Status != constant.StatusUnavailable || status.Database != constant.DatabaseDown {
		t.Fatalf("status = %#v", status)
	}
}

func TestVersion(t *testing.T) {
	svc := business.NewHealthService(mockHealthRepo{}, business.BuildInfo{
		Version:   "1.0.0",
		Commit:    "abc",
		BuildTime: "now",
	})
	v := svc.Version()
	if v.Version != "1.0.0" || v.Commit != "abc" || v.BuildTime != "now" {
		t.Fatalf("version = %#v", v)
	}
}

package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// the store_test_suite spins up 1 real postgres container for the whole test file,
// running migrations once, rather than pre-test (much faster)
type StoreTestSuite struct {
	suite.Suite
	store     *Store
	container *tcpostgres.PostgresContainer
}

func (s *StoreTestSuite) SetupSuite() {
	ctx := context.Background()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("cage_test"),
		tcpostgres.WithUsername("cage"),
		tcpostgres.WithPassword("cage"),
	)
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(ctx)
	require.NoError(s.T(), err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(s.T(), err)

	// force to ipv4 explicitly
	if host == "localhost" {
		host = "127.0.0.1"
	}

	connStr := fmt.Sprintf("postgres://cage:cage@%s:%s/cage_test?sslmode=disable", host, port.Port())
	require.NoError(s.T(), err)

	// run migrations against this fetch throwaway db
	err = runTestMigrations(connStr)
	require.NoError(s.T(), err)

	st, err := NewStore(ctx, connStr)
	require.NoError(s.T(), err)
	s.store = st
}

func (s *StoreTestSuite) TearDownSuite() {
	_ = s.container.Terminate(context.Background())
}

func TestStoreSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainers suite in short mode")
	}
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) TestSaveAndGet() {
	ctx := context.Background()
	id := uuid.NewString()
	sb := &Sandbox{
		ID:           id,
		ContainerID:  "container-1",
		Status:       StatusRunning,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		TemplateSlug: "base",
	}

	err := s.store.Save(ctx, sb)
	s.Require().NoError(err)

	got, err := s.store.Get(ctx, id)
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(sb.ContainerID, got.ContainerID)
	s.Equal(sb.Status, got.Status)
}

func (s *StoreTestSuite) TestGet_NotFound() {
	got, err := s.store.Get(context.Background(), uuid.NewString())
	s.Require().NoError(err)
	s.Nil(got, "should return nil, nil for a missing sandbox, not an error")
}

func (s *StoreTestSuite) TestDelete() {
	ctx := context.Background()
	id := uuid.NewString()
	sb := &Sandbox{
		ID:          id,
		ContainerID: "c-1",
		Status:      StatusRunning,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	s.Require().NoError(s.store.Save(ctx, sb))

	s.Require().NoError(s.store.Delete(ctx, id))

	got, err := s.store.Get(ctx, id)
	s.Require().NoError(err)
	s.Nil(got)
}

func (s *StoreTestSuite) TestListExpired() {
	ctx := context.Background()
	id := uuid.NewString()
	expired := &Sandbox{
		ID:          id,
		ContainerID: "c-2",
		Status:      StatusRunning,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // expired already
	}
	s.Require().NoError(s.store.Save(ctx, expired))

	results, err := s.store.ListExpired(ctx)
	s.Require().NoError(err)

	found := false
	for _, sb := range results {
		if sb.ID == id {
			found = true
		}
	}

	s.True(found, "expired sandbox should appear in ListExpired results")
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shop/backend/internal/domain"
)

var (
	errBoom         = errors.New("boom")
	errCommitFailed = errors.New("commit failed")
)

type mockProductRepo struct {
	createFn   func(ctx context.Context, product *domain.Product) error
	getAllFn   func(ctx context.Context, page, pageSize int) ([]domain.Product, error)
	getByIDFn  func(ctx context.Context, id uint) (*domain.Product, error)
	getByIDsFn func(ctx context.Context, ids []uint) ([]domain.Product, error)
	updateFn   func(ctx context.Context, product *domain.Product) error
	deleteFn   func(ctx context.Context, id uint) error

	createCalls   int
	getAllCalls   int
	getByIDCalls  int
	getByIDsCalls int
	updateCalls   int
	deleteCalls   int
	lastIDs       []uint
	lastGetByID   uint
}

var _ ProductRepository = (*mockProductRepo)(nil)

func (m *mockProductRepo) Create(ctx context.Context, product *domain.Product) error {
	m.createCalls++
	if m.createFn == nil {
		panic("not implemented")
	}
	return m.createFn(ctx, product)
}

func (m *mockProductRepo) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
	m.getAllCalls++
	if m.getAllFn == nil {
		panic("not implemented")
	}
	return m.getAllFn(ctx, page, pageSize)
}

func (m *mockProductRepo) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	m.getByIDCalls++
	m.lastGetByID = id
	if m.getByIDFn == nil {
		panic("not implemented")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockProductRepo) GetByIDs(ctx context.Context, ids []uint) ([]domain.Product, error) {
	m.getByIDsCalls++
	m.lastIDs = ids
	if m.getByIDsFn == nil {
		panic("not implemented")
	}
	return m.getByIDsFn(ctx, ids)
}

func (m *mockProductRepo) Update(ctx context.Context, product *domain.Product) error {
	m.updateCalls++
	if m.updateFn == nil {
		panic("not implemented")
	}
	return m.updateFn(ctx, product)
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint) error {
	m.deleteCalls++
	if m.deleteFn == nil {
		panic("not implemented")
	}
	return m.deleteFn(ctx, id)
}

func (m *mockProductRepo) GetByName(ctx context.Context, name string) (*domain.Product, error) {
	panic("not implemented")
}

func (m *mockProductRepo) ExistsInOrders(ctx context.Context, id uint) (bool, error) {
	panic("not implemented")
}

type mockTxManager struct {
	beginErr  error
	commitErr error

	executed   int
	committed  int
	rolledBack int
}

var _ TransactionManager = (*mockTxManager)(nil)

func (m *mockTxManager) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	m.executed++
	if m.beginErr != nil {
		return m.beginErr
	}
	if err := fn(ctx); err != nil {
		m.rolledBack++
		return err
	}
	if m.commitErr != nil {
		m.rolledBack++
		return m.commitErr
	}
	m.committed++
	return nil
}

func assertValidationError(t *testing.T, err error, wantField string) {
	t.Helper()
	var vErr *domain.ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("errors.As(*domain.ValidationError) = false, got %T: %v", err, err)
	}
	if vErr.Field != wantField {
		t.Errorf("validation field = %q, want %q", vErr.Field, wantField)
	}
}

func TestProductService_Create(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	var storedFromDB *domain.Product

	tests := []struct {
		name                string
		input               *domain.Product
		repo                func() *mockProductRepo
		beginErr            error
		commitErr           error
		wantValidationField string
		wantErrIs           error
		wantCreateCalls     int
		wantGetByIDCalls    int
		wantTxExecuted      int
		wantID              uint
		wantSameAsGetByID   bool
	}{
		{
			name:                "empty name is rejected before transaction",
			input:               &domain.Product{Name: "", Price: decimal.NewFromInt(10)},
			wantValidationField: "name",
		},
		{
			name:                "zero price is rejected before transaction",
			input:               &domain.Product{Name: "Keyboard", Price: decimal.Zero},
			wantValidationField: "price",
		},
		{
			name:                "negative price is rejected before transaction",
			input:               &domain.Product{Name: "Keyboard", Price: decimal.NewFromInt(-5)},
			wantValidationField: "price",
		},
		{
			name:  "valid product is created and reloaded",
			input: &domain.Product{Name: "Keyboard", Price: decimal.NewFromInt(10)},
			repo: func() *mockProductRepo {
				storedFromDB = &domain.Product{
					ID:        7,
					Name:      "Keyboard",
					Price:     decimal.NewFromInt(10),
					CreatedAt: now,
					UpdatedAt: now,
				}
				return &mockProductRepo{
					createFn: func(ctx context.Context, p *domain.Product) error {
						p.ID = 7
						p.CreatedAt = now
						p.UpdatedAt = now
						return nil
					},
					getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
						return storedFromDB, nil
					},
				}
			},
			wantCreateCalls:   1,
			wantGetByIDCalls:  1,
			wantTxExecuted:    1,
			wantID:            7,
			wantSameAsGetByID: true,
		},
		{
			name:  "repository conflict is propagated and transaction rolls back",
			input: &domain.Product{Name: "Keyboard", Price: decimal.NewFromInt(10)},
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					createFn: func(ctx context.Context, p *domain.Product) error {
						return domain.ErrAlreadyExists
					},
				}
			},
			wantErrIs:       domain.ErrAlreadyExists,
			wantCreateCalls: 1,
			wantTxExecuted:  1,
		},
		{
			name:           "transaction start failure is propagated",
			input:          &domain.Product{Name: "Keyboard", Price: decimal.NewFromInt(10)},
			beginErr:       errBoom,
			wantErrIs:      errBoom,
			wantTxExecuted: 1,
		},
		{
			name:  "commit failure is propagated and rolls back",
			input: &domain.Product{Name: "Keyboard", Price: decimal.NewFromInt(10)},
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					createFn: func(ctx context.Context, p *domain.Product) error {
						p.ID = 7
						return nil
					},
					getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
						return &domain.Product{ID: id}, nil
					},
				}
			},
			commitErr:        errCommitFailed,
			wantErrIs:        errCommitFailed,
			wantCreateCalls:  1,
			wantGetByIDCalls: 1,
			wantTxExecuted:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedFromDB = nil

			repo := &mockProductRepo{}
			if tt.repo != nil {
				repo = tt.repo()
			}
			tx := &mockTxManager{beginErr: tt.beginErr, commitErr: tt.commitErr}

			svc := NewProductService(repo, tx)
			got, err := svc.Create(context.Background(), tt.input)

			switch {
			case tt.wantValidationField != "":
				assertValidationError(t, err, tt.wantValidationField)
			case tt.wantErrIs != nil:
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("errors.Is(err, %v) = false, got %v", tt.wantErrIs, err)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if err != nil && got != nil {
				t.Errorf("got = %+v, want nil on error", got)
			}
			if repo.createCalls != tt.wantCreateCalls {
				t.Errorf("repo.Create calls = %d, want %d", repo.createCalls, tt.wantCreateCalls)
			}
			if repo.getByIDCalls != tt.wantGetByIDCalls {
				t.Errorf("repo.GetByID calls = %d, want %d", repo.getByIDCalls, tt.wantGetByIDCalls)
			}
			if tx.executed != tt.wantTxExecuted {
				t.Errorf("tx.Execute calls = %d, want %d", tx.executed, tt.wantTxExecuted)
			}
			if err != nil && tx.executed > 0 && tt.beginErr == nil && tx.rolledBack != 1 {
				t.Errorf("tx.rolledBack = %d, want 1", tx.rolledBack)
			}
			if err != nil && tx.committed != 0 {
				t.Errorf("tx.committed = %d, want 0 on error", tx.committed)
			}
			if err == nil && tx.executed > 0 && tx.committed != 1 {
				t.Errorf("tx.committed = %d, want 1", tx.committed)
			}
			if tt.wantSameAsGetByID && got != storedFromDB {
				t.Errorf("service must return the object loaded by repo.GetByID, not the mutated input")
			}
			if tt.wantID != 0 && (got == nil || got.ID != tt.wantID) {
				t.Errorf("got.ID = %v, want %d", got, tt.wantID)
			}
		})
	}
}

func TestProductService_Update(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	var storedFromDB *domain.Product

	tests := []struct {
		name                string
		input               *domain.Product
		repo                func() *mockProductRepo
		beginErr            error
		commitErr           error
		wantValidationField string
		wantErrIs           error
		wantUpdateCalls     int
		wantGetByIDCalls    int
		wantTxExecuted      int
		wantTxCommitted     int
		wantID              uint
		wantSameAsGetByID   bool
	}{
		{
			name:                "zero id is rejected before transaction",
			input:               &domain.Product{ID: 0, Name: "Keyboard", Price: decimal.NewFromInt(10)},
			wantValidationField: "id",
		},
		{
			name:                "empty name is rejected",
			input:               &domain.Product{ID: 7, Name: "", Price: decimal.NewFromInt(10)},
			wantValidationField: "name",
		},
		{
			name:                "zero price is rejected",
			input:               &domain.Product{ID: 7, Name: "Keyboard", Price: decimal.Zero},
			wantValidationField: "price",
		},
		{
			name:  "missing product propagates ErrNotFound",
			input: &domain.Product{ID: 404, Name: "Keyboard", Price: decimal.NewFromInt(10)},
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					updateFn: func(ctx context.Context, p *domain.Product) error {
						return domain.ErrNotFound
					},
				}
			},
			wantErrIs:       domain.ErrNotFound,
			wantUpdateCalls: 1,
			wantTxExecuted:  1,
			wantTxCommitted: 0,
		},
		{
			name:  "valid update reloads product",
			input: &domain.Product{ID: 7, Name: "Keyboard Pro", Price: decimal.NewFromInt(20)},
			repo: func() *mockProductRepo {
				storedFromDB = &domain.Product{
					ID:        7,
					Name:      "Keyboard Pro",
					Price:     decimal.NewFromInt(20),
					CreatedAt: now,
					UpdatedAt: now,
				}
				return &mockProductRepo{
					updateFn: func(ctx context.Context, p *domain.Product) error {
						p.UpdatedAt = now
						return nil
					},
					getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
						return storedFromDB, nil
					},
				}
			},
			wantUpdateCalls:   1,
			wantGetByIDCalls:  1,
			wantTxExecuted:    1,
			wantTxCommitted:   1,
			wantID:            7,
			wantSameAsGetByID: true,
		},
		{
			name:           "transaction start failure is propagated",
			input:          &domain.Product{ID: 7, Name: "Keyboard Pro", Price: decimal.NewFromInt(20)},
			beginErr:       errBoom,
			wantErrIs:      errBoom,
			wantTxExecuted: 1,
			wantTxCommitted: 0,
		},
		{
			name:  "commit failure is propagated and rolls back",
			input: &domain.Product{ID: 7, Name: "Keyboard Pro", Price: decimal.NewFromInt(20)},
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					updateFn: func(ctx context.Context, p *domain.Product) error {
						p.UpdatedAt = now
						return nil
					},
					getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
						return &domain.Product{ID: id}, nil
					},
				}
			},
			commitErr:        errCommitFailed,
			wantErrIs:        errCommitFailed,
			wantUpdateCalls:  1,
			wantGetByIDCalls: 1,
			wantTxExecuted:   1,
			wantTxCommitted:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedFromDB = nil

			repo := &mockProductRepo{}
			if tt.repo != nil {
				repo = tt.repo()
			}
			tx := &mockTxManager{beginErr: tt.beginErr, commitErr: tt.commitErr}

			svc := NewProductService(repo, tx)
			got, err := svc.Update(context.Background(), tt.input)

			switch {
			case tt.wantValidationField != "":
				assertValidationError(t, err, tt.wantValidationField)
			case tt.wantErrIs != nil:
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("errors.Is(err, %v) = false, got %v", tt.wantErrIs, err)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if repo.updateCalls != tt.wantUpdateCalls {
				t.Errorf("repo.Update calls = %d, want %d", repo.updateCalls, tt.wantUpdateCalls)
			}
			if repo.getByIDCalls != tt.wantGetByIDCalls {
				t.Errorf("repo.GetByID calls = %d, want %d", repo.getByIDCalls, tt.wantGetByIDCalls)
			}
			if tt.wantValidationField != "" && tx.executed != 0 {
				t.Errorf("tx.Execute calls = %d, want 0 on validation failure", tx.executed)
			}
			if tx.executed != tt.wantTxExecuted {
				t.Errorf("tx.Execute calls = %d, want %d", tx.executed, tt.wantTxExecuted)
			}
			if tx.committed != tt.wantTxCommitted {
				t.Errorf("tx.committed = %d, want %d", tx.committed, tt.wantTxCommitted)
			}
			if err != nil && tx.executed > 0 && tx.rolledBack != 1 && tt.beginErr == nil {
				t.Errorf("tx.rolledBack = %d, want 1", tx.rolledBack)
			}
			if tt.wantSameAsGetByID && got != storedFromDB {
				t.Errorf("service must return the object loaded by repo.GetByID")
			}
			if tt.wantID != 0 && (got == nil || got.ID != tt.wantID) {
				t.Errorf("got.ID = %v, want %d", got, tt.wantID)
			}
		})
	}
}

func TestProductService_Delete(t *testing.T) {
	tests := []struct {
		name                string
		id                  uint
		repo                func() *mockProductRepo
		wantValidationField string
		wantErrIs           error
		wantDeleteCalls     int
		wantTxExecuted      int
	}{
		{
			name:                "zero id is rejected before transaction",
			id:                  0,
			wantValidationField: "id",
		},
		{
			name: "product in use propagates ErrInUse",
			id:   7,
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					deleteFn: func(ctx context.Context, id uint) error {
						return domain.ErrInUse
					},
				}
			},
			wantErrIs:       domain.ErrInUse,
			wantDeleteCalls: 1,
			wantTxExecuted:  1,
		},
		{
			name: "missing product propagates ErrNotFound",
			id:   404,
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					deleteFn: func(ctx context.Context, id uint) error {
						return domain.ErrNotFound
					},
				}
			},
			wantErrIs:       domain.ErrNotFound,
			wantDeleteCalls: 1,
			wantTxExecuted:  1,
		},
		{
			name: "successful delete commits",
			id:   7,
			repo: func() *mockProductRepo {
				return &mockProductRepo{
					deleteFn: func(ctx context.Context, id uint) error {
						return nil
					},
				}
			},
			wantDeleteCalls: 1,
			wantTxExecuted:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepo{}
			if tt.repo != nil {
				repo = tt.repo()
			}
			tx := &mockTxManager{}

			svc := NewProductService(repo, tx)
			err := svc.Delete(context.Background(), tt.id)

			switch {
			case tt.wantValidationField != "":
				assertValidationError(t, err, tt.wantValidationField)
			case tt.wantErrIs != nil:
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("errors.Is(err, %v) = false, got %v", tt.wantErrIs, err)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if repo.deleteCalls != tt.wantDeleteCalls {
				t.Errorf("repo.Delete calls = %d, want %d", repo.deleteCalls, tt.wantDeleteCalls)
			}
			if tx.executed != tt.wantTxExecuted {
				t.Errorf("tx.Execute calls = %d, want %d", tx.executed, tt.wantTxExecuted)
			}
			if err != nil && tx.executed > 0 && tx.rolledBack != 1 {
				t.Errorf("tx.rolledBack = %d, want 1", tx.rolledBack)
			}
			if err == nil && tx.committed != 1 {
				t.Errorf("tx.committed = %d, want 1", tx.committed)
			}
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	t.Run("zero id is rejected", func(t *testing.T) {
		tx := &mockTxManager{}
		repo := &mockProductRepo{}
		svc := NewProductService(repo, tx)

		got, err := svc.GetByID(context.Background(), 0)

		assertValidationError(t, err, "id")
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
		if repo.getByIDCalls != 0 {
			t.Errorf("repo.GetByID calls = %d, want 0", repo.getByIDCalls)
		}
	})

	t.Run("missing product propagates ErrNotFound", func(t *testing.T) {
		repo := &mockProductRepo{
			getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
				return nil, domain.ErrNotFound
			},
		}
		svc := NewProductService(repo, &mockTxManager{})

		got, err := svc.GetByID(context.Background(), 404)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("errors.Is(err, ErrNotFound) = false, got %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
	})

	t.Run("found product is returned", func(t *testing.T) {
		want := &domain.Product{ID: 7, Name: "Keyboard", Price: decimal.NewFromInt(10)}
		repo := &mockProductRepo{
			getByIDFn: func(ctx context.Context, id uint) (*domain.Product, error) {
				return want, nil
			},
		}
		svc := NewProductService(repo, &mockTxManager{})

		got, err := svc.GetByID(context.Background(), 7)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got = %+v, want %+v", got, want)
		}
	})
}

func TestProductService_GetAll(t *testing.T) {
	t.Run("repository error is wrapped and unwrappable", func(t *testing.T) {
		repo := &mockProductRepo{
			getAllFn: func(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
				return nil, errBoom
			},
		}
		svc := NewProductService(repo, &mockTxManager{})

		got, err := svc.GetAll(context.Background(), 1, 20)

		if !errors.Is(err, errBoom) {
			t.Fatalf("errors.Is(err, errBoom) = false, got %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
	})

	t.Run("products are returned without transaction", func(t *testing.T) {
		tx := &mockTxManager{}
		repo := &mockProductRepo{
			getAllFn: func(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
				return []domain.Product{
					{ID: 1, Name: "Keyboard", Price: decimal.NewFromInt(10)},
					{ID: 2, Name: "Mouse", Price: decimal.NewFromInt(5)},
				}, nil
			},
		}
		svc := NewProductService(repo, tx)

		got, err := svc.GetAll(context.Background(), 2, 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, want 2", len(got))
		}
		if repo.getAllCalls != 1 {
			t.Errorf("repo.GetAll calls = %d, want 1", repo.getAllCalls)
		}
		if tx.executed != 0 {
			t.Errorf("tx.Execute calls = %d, want 0 for read-only path", tx.executed)
		}
	})
}
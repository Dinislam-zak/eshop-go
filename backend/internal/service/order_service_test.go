package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/shop/backend/internal/domain"
)

type mockOrderRepo struct {
	createFn   func(ctx context.Context, order *domain.Order) error
	getAllFn   func(ctx context.Context, page, pageSize int) ([]domain.Order, error)
	getByIDFn  func(ctx context.Context, id uint) (*domain.Order, error)
	updateFn   func(ctx context.Context, order *domain.Order) error
	deleteFn   func(ctx context.Context, id uint) error

	createCalls  int
	getAllCalls  int
	getByIDCalls int
	updateCalls  int
	deleteCalls  int
	lastGetByID  uint
}

var _ OrderRepository = (*mockOrderRepo)(nil)

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	m.createCalls++
	if m.createFn == nil {
		panic("not implemented")
	}
	return m.createFn(ctx, order)
}

func (m *mockOrderRepo) GetAll(ctx context.Context, page, pageSize int) ([]domain.Order, error) {
	m.getAllCalls++
	if m.getAllFn == nil {
		panic("not implemented")
	}
	return m.getAllFn(ctx, page, pageSize)
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	m.getByIDCalls++
	m.lastGetByID = id
	if m.getByIDFn == nil {
		panic("not implemented")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockOrderRepo) GetByNumber(ctx context.Context, number string) (*domain.Order, error) {
	panic("not implemented")
}

func (m *mockOrderRepo) Update(ctx context.Context, order *domain.Order) error {
	m.updateCalls++
	if m.updateFn == nil {
		panic("not implemented")
	}
	return m.updateFn(ctx, order)
}

func (m *mockOrderRepo) Delete(ctx context.Context, id uint) error {
	m.deleteCalls++
	if m.deleteFn == nil {
		panic("not implemented")
	}
	return m.deleteFn(ctx, id)
}

func productsForIDs(ids []uint) []domain.Product {
	out := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		out = append(out, domain.Product{
			ID:    id,
			Name:  "Product",
			Price: decimal.NewFromInt(10),
		})
	}
	return out
}

func TestOrderService_Create_Validation(t *testing.T) {
	validItems := []domain.OrderItem{{ProductID: 1, Quantity: 2}}

	tests := []struct {
		name      string
		input     *domain.Order
		wantField string
	}{
		{
			name:      "empty number",
			input:     &domain.Order{Number: "", DeliveryAddress: "Nevsky 1", Items: validItems},
			wantField: "number",
		},
		{
			name:      "empty delivery address",
			input:     &domain.Order{Number: "A-1", DeliveryAddress: "", Items: validItems},
			wantField: "delivery_address",
		},
		{
			name:      "empty items",
			input:     &domain.Order{Number: "A-1", DeliveryAddress: "Nevsky 1", Items: nil},
			wantField: "items",
		},
		{
			name: "zero quantity",
			input: &domain.Order{
				Number:          "A-1",
				DeliveryAddress: "Nevsky 1",
				Items:           []domain.OrderItem{{ProductID: 1, Quantity: 0}},
			},
			wantField: "items.quantity",
		},
		{
			name: "negative quantity",
			input: &domain.Order{
				Number:          "A-1",
				DeliveryAddress: "Nevsky 1",
				Items:           []domain.OrderItem{{ProductID: 1, Quantity: -3}},
			},
			wantField: "items.quantity",
		},
		{
			name: "duplicate product id",
			input: &domain.Order{
				Number:          "A-1",
				DeliveryAddress: "Nevsky 1",
				Items: []domain.OrderItem{
					{ProductID: 1, Quantity: 1},
					{ProductID: 1, Quantity: 2},
				},
			},
			wantField: "items.product_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &mockProductRepo{}
			orderRepo := &mockOrderRepo{}
			tx := &mockTxManager{}

			svc := NewOrderService(orderRepo, productRepo, tx)
			got, err := svc.Create(context.Background(), tt.input)

			assertValidationError(t, err, tt.wantField)
			if got != nil {
				t.Errorf("got = %+v, want nil", got)
			}
			if tx.executed != 0 {
				t.Errorf("tx.Execute calls = %d, want 0 before validation passes", tx.executed)
			}
			if productRepo.getByIDsCalls != 0 {
				t.Errorf("productRepo.GetByIDs calls = %d, want 0", productRepo.getByIDsCalls)
			}
			if orderRepo.createCalls != 0 {
				t.Errorf("orderRepo.Create calls = %d, want 0", orderRepo.createCalls)
			}
		})
	}
}

func TestOrderService_Create_CallsGetByIDsOnceWithFullList(t *testing.T) {
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
			return productsForIDs(ids), nil
		},
	}
	orderRepo := &mockOrderRepo{
		createFn: func(ctx context.Context, o *domain.Order) error {
			o.ID = 11
			return nil
		},
		getByIDFn: func(ctx context.Context, id uint) (*domain.Order, error) {
			return &domain.Order{
				ID:     id,
				Number: "A-1",
				Items: []domain.OrderItem{
					{ProductID: 1, Quantity: 2, Product: domain.Product{ID: 1, Name: "Keyboard", Price: decimal.NewFromInt(10)}},
					{ProductID: 5, Quantity: 1, Product: domain.Product{ID: 5, Name: "Mouse", Price: decimal.NewFromInt(5)}},
					{ProductID: 9, Quantity: 3, Product: domain.Product{ID: 9, Name: "Pad", Price: decimal.NewFromInt(3)}},
				},
			}, nil
		},
	}
	tx := &mockTxManager{}

	svc := NewOrderService(orderRepo, productRepo, tx)
	got, err := svc.Create(context.Background(), &domain.Order{
		Number:          "A-1",
		DeliveryAddress: "Nevsky 1",
		Items: []domain.OrderItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 5, Quantity: 1},
			{ProductID: 9, Quantity: 3},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if productRepo.getByIDsCalls != 1 {
		t.Errorf("productRepo.GetByIDs calls = %d, want exactly 1", productRepo.getByIDsCalls)
	}
	if !slices.Equal(productRepo.lastIDs, []uint{1, 5, 9}) {
		t.Errorf("productRepo.GetByIDs ids = %v, want [1 5 9]", productRepo.lastIDs)
	}
	if orderRepo.getByIDCalls != 1 || orderRepo.lastGetByID != 11 {
		t.Errorf("orderRepo.GetByID calls = %d (id %d), want 1 (id 11)", orderRepo.getByIDCalls, orderRepo.lastGetByID)
	}
	if tx.committed != 1 {
		t.Errorf("tx.committed = %d, want 1", tx.committed)
	}
	if len(got.Items) != 3 {
		t.Fatalf("len(got.Items) = %d, want 3", len(got.Items))
	}
	for i, item := range got.Items {
		if item.Product.ID == 0 || item.Product.Name == "" {
			t.Errorf("got.Items[%d].Product is not populated: %+v", i, item.Product)
		}
	}
}

func TestOrderService_Create_MissingProduct(t *testing.T) {
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
			return productsForIDs(ids[:len(ids)-1]), nil
		},
	}
	orderRepo := &mockOrderRepo{}
	tx := &mockTxManager{}

	svc := NewOrderService(orderRepo, productRepo, tx)
	got, err := svc.Create(context.Background(), &domain.Order{
		Number:          "A-1",
		DeliveryAddress: "Nevsky 1",
		Items: []domain.OrderItem{
			{ProductID: 1, Quantity: 1},
			{ProductID: 404, Quantity: 1},
		},
	})

	assertValidationError(t, err, "items.product_id")
	if got != nil {
		t.Errorf("got = %+v, want nil", got)
	}
	if orderRepo.createCalls != 0 {
		t.Errorf("orderRepo.Create calls = %d, want 0", orderRepo.createCalls)
	}
	if tx.rolledBack != 1 {
		t.Errorf("tx.rolledBack = %d, want 1", tx.rolledBack)
	}
}

func TestOrderService_Create_RollsBackOnRepoError(t *testing.T) {
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
			return productsForIDs(ids), nil
		},
	}
	orderRepo := &mockOrderRepo{
		createFn: func(ctx context.Context, o *domain.Order) error {
			return domain.ErrAlreadyExists
		},
	}
	tx := &mockTxManager{}

	svc := NewOrderService(orderRepo, productRepo, tx)
	got, err := svc.Create(context.Background(), &domain.Order{
		Number:          "A-1",
		DeliveryAddress: "Nevsky 1",
		Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("errors.Is(err, ErrAlreadyExists) = false, got %v", err)
	}
	if got != nil {
		t.Errorf("got = %+v, want nil", got)
	}
	if orderRepo.getByIDCalls != 0 {
		t.Errorf("orderRepo.GetByID calls = %d, want 0 after failed create", orderRepo.getByIDCalls)
	}
	if tx.rolledBack != 1 || tx.committed != 0 {
		t.Errorf("tx.rolledBack = %d, tx.committed = %d, want 1 and 0", tx.rolledBack, tx.committed)
	}
}

func TestOrderService_Create_ReloadFailureIsVisible(t *testing.T) {
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
			return productsForIDs(ids), nil
		},
	}
	orderRepo := &mockOrderRepo{
		createFn: func(ctx context.Context, o *domain.Order) error {
			o.ID = 11
			return nil
		},
		getByIDFn: func(ctx context.Context, id uint) (*domain.Order, error) {
			return nil, domain.ErrNotFound
		},
	}
	tx := &mockTxManager{}

	svc := NewOrderService(orderRepo, productRepo, tx)
	got, err := svc.Create(context.Background(), &domain.Order{
		Number:          "A-1",
		DeliveryAddress: "Nevsky 1",
		Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
	})

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false, got %v", err)
	}
	if got != nil {
		t.Errorf("got = %+v, want nil", got)
	}
	if tx.rolledBack != 1 || tx.committed != 0 {
		t.Errorf("tx.rolledBack = %d, tx.committed = %d, want 1 and 0", tx.rolledBack, tx.committed)
	}
}

func TestOrderService_Create_ProductLookupErrorIsUnwrappable(t *testing.T) {
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
			return nil, errBoom
		},
	}
	orderRepo := &mockOrderRepo{}
	tx := &mockTxManager{}

	svc := NewOrderService(orderRepo, productRepo, tx)
	got, err := svc.Create(context.Background(), &domain.Order{
		Number:          "A-1",
		DeliveryAddress: "Nevsky 1",
		Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
	})

	if !errors.Is(err, errBoom) {
		t.Fatalf("errors.Is(err, errBoom) = false, got %v", err)
	}
	if got != nil {
		t.Errorf("got = %+v, want nil", got)
	}
	if orderRepo.createCalls != 0 {
		t.Errorf("orderRepo.Create calls = %d, want 0", orderRepo.createCalls)
	}
}

func TestOrderService_Update(t *testing.T) {
	t.Run("zero id is rejected before transaction", func(t *testing.T) {
		productRepo := &mockProductRepo{}
		orderRepo := &mockOrderRepo{}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, productRepo, tx)
		got, err := svc.Update(context.Background(), &domain.Order{
			ID:              0,
			Number:          "A-1",
			DeliveryAddress: "Nevsky 1",
			Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
		})

		assertValidationError(t, err, "id")
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
		if tx.executed != 0 {
			t.Errorf("tx.Execute calls = %d, want 0", tx.executed)
		}
	})

	t.Run("validation is shared with create", func(t *testing.T) {
		tx := &mockTxManager{}
		svc := NewOrderService(&mockOrderRepo{}, &mockProductRepo{}, tx)

		got, err := svc.Update(context.Background(), &domain.Order{
			ID:              11,
			Number:          "A-1",
			DeliveryAddress: "",
			Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
		})

		assertValidationError(t, err, "delivery_address")
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
		if tx.executed != 0 {
			t.Errorf("tx.Execute calls = %d, want 0", tx.executed)
		}
	})

	t.Run("missing order propagates ErrNotFound", func(t *testing.T) {
		productRepo := &mockProductRepo{
			getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
				return productsForIDs(ids), nil
			},
		}
		orderRepo := &mockOrderRepo{
			updateFn: func(ctx context.Context, o *domain.Order) error {
				return domain.ErrNotFound
			},
		}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, productRepo, tx)
		got, err := svc.Update(context.Background(), &domain.Order{
			ID:              404,
			Number:          "A-1",
			DeliveryAddress: "Nevsky 1",
			Items:           []domain.OrderItem{{ProductID: 1, Quantity: 1}},
		})

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("errors.Is(err, ErrNotFound) = false, got %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
		if tx.rolledBack != 1 {
			t.Errorf("tx.rolledBack = %d, want 1", tx.rolledBack)
		}
	})

	t.Run("valid update reloads order with nested products", func(t *testing.T) {
		productRepo := &mockProductRepo{
			getByIDsFn: func(ctx context.Context, ids []uint) ([]domain.Product, error) {
				return productsForIDs(ids), nil
			},
		}
		orderRepo := &mockOrderRepo{
			updateFn: func(ctx context.Context, o *domain.Order) error {
				return nil
			},
			getByIDFn: func(ctx context.Context, id uint) (*domain.Order, error) {
				return &domain.Order{
					ID:              id,
					Number:          "A-2",
					DeliveryAddress: "Nevsky 2",
					Items: []domain.OrderItem{
						{ProductID: 1, Quantity: 4, Product: domain.Product{ID: 1, Name: "Keyboard", Price: decimal.NewFromInt(10)}},
					},
				}, nil
			},
		}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, productRepo, tx)
		got, err := svc.Update(context.Background(), &domain.Order{
			ID:              11,
			Number:          "A-2",
			DeliveryAddress: "Nevsky 2",
			Items:           []domain.OrderItem{{ProductID: 1, Quantity: 4}},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if productRepo.getByIDsCalls != 1 {
			t.Errorf("productRepo.GetByIDs calls = %d, want 1", productRepo.getByIDsCalls)
		}
		if orderRepo.getByIDCalls != 1 || orderRepo.lastGetByID != 11 {
			t.Errorf("orderRepo.GetByID calls = %d (id %d), want 1 (id 11)", orderRepo.getByIDCalls, orderRepo.lastGetByID)
		}
		if got.Number != "A-2" || len(got.Items) != 1 || got.Items[0].Product.ID != 1 {
			t.Errorf("got = %+v, want reloaded order with nested product", got)
		}
		if tx.committed != 1 {
			t.Errorf("tx.committed = %d, want 1", tx.committed)
		}
	})
}

func TestOrderService_Delete(t *testing.T) {
	t.Run("zero id is rejected before transaction", func(t *testing.T) {
		orderRepo := &mockOrderRepo{}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, &mockProductRepo{}, tx)
		err := svc.Delete(context.Background(), 0)

		assertValidationError(t, err, "id")
		if orderRepo.deleteCalls != 0 {
			t.Errorf("orderRepo.Delete calls = %d, want 0", orderRepo.deleteCalls)
		}
		if tx.executed != 0 {
			t.Errorf("tx.Execute calls = %d, want 0", tx.executed)
		}
	})

	t.Run("missing order propagates ErrNotFound and rolls back", func(t *testing.T) {
		orderRepo := &mockOrderRepo{
			deleteFn: func(ctx context.Context, id uint) error {
				return domain.ErrNotFound
			},
		}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, &mockProductRepo{}, tx)
		err := svc.Delete(context.Background(), 404)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("errors.Is(err, ErrNotFound) = false, got %v", err)
		}
		if tx.rolledBack != 1 || tx.committed != 0 {
			t.Errorf("tx.rolledBack = %d, tx.committed = %d, want 1 and 0", tx.rolledBack, tx.committed)
		}
	})

	t.Run("successful delete commits once", func(t *testing.T) {
		orderRepo := &mockOrderRepo{
			deleteFn: func(ctx context.Context, id uint) error {
				return nil
			},
		}
		tx := &mockTxManager{}

		svc := NewOrderService(orderRepo, &mockProductRepo{}, tx)
		err := svc.Delete(context.Background(), 11)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if orderRepo.deleteCalls != 1 {
			t.Errorf("orderRepo.Delete calls = %d, want 1", orderRepo.deleteCalls)
		}
		if tx.executed != 1 || tx.committed != 1 {
			t.Errorf("tx.executed = %d, tx.committed = %d, want 1 and 1", tx.executed, tx.committed)
		}
	})
}

func TestOrderService_GetByID(t *testing.T) {
	t.Run("zero id is rejected", func(t *testing.T) {
		orderRepo := &mockOrderRepo{}
		svc := NewOrderService(orderRepo, &mockProductRepo{}, &mockTxManager{})

		got, err := svc.GetByID(context.Background(), 0)

		assertValidationError(t, err, "id")
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
		if orderRepo.getByIDCalls != 0 {
			t.Errorf("orderRepo.GetByID calls = %d, want 0", orderRepo.getByIDCalls)
		}
	})

	t.Run("missing order propagates ErrNotFound", func(t *testing.T) {
		orderRepo := &mockOrderRepo{
			getByIDFn: func(ctx context.Context, id uint) (*domain.Order, error) {
				return nil, domain.ErrNotFound
			},
		}
		svc := NewOrderService(orderRepo, &mockProductRepo{}, &mockTxManager{})

		got, err := svc.GetByID(context.Background(), 404)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("errors.Is(err, ErrNotFound) = false, got %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
	})

	t.Run("order is returned", func(t *testing.T) {
		want := &domain.Order{ID: 11, Number: "A-1", DeliveryAddress: "Nevsky 1"}
		orderRepo := &mockOrderRepo{
			getByIDFn: func(ctx context.Context, id uint) (*domain.Order, error) {
				return want, nil
			},
		}
		svc := NewOrderService(orderRepo, &mockProductRepo{}, &mockTxManager{})

		got, err := svc.GetByID(context.Background(), 11)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got = %+v, want %+v", got, want)
		}
	})
}

func TestOrderService_GetAll(t *testing.T) {
	t.Run("repository error is unwrappable", func(t *testing.T) {
		orderRepo := &mockOrderRepo{
			getAllFn: func(ctx context.Context, page, pageSize int) ([]domain.Order, error) {
				return nil, errBoom
			},
		}
		svc := NewOrderService(orderRepo, &mockProductRepo{}, &mockTxManager{})

		got, err := svc.GetAll(context.Background(), 1, 20)

		if !errors.Is(err, errBoom) {
			t.Fatalf("errors.Is(err, errBoom) = false, got %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
	})

	t.Run("orders are returned without transaction", func(t *testing.T) {
		tx := &mockTxManager{}
		orderRepo := &mockOrderRepo{
			getAllFn: func(ctx context.Context, page, pageSize int) ([]domain.Order, error) {
				return []domain.Order{{ID: 11}, {ID: 12}}, nil
			},
		}
		svc := NewOrderService(orderRepo, &mockProductRepo{}, tx)

		got, err := svc.GetAll(context.Background(), 1, 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, want 2", len(got))
		}
		if orderRepo.getAllCalls != 1 {
			t.Errorf("orderRepo.GetAll calls = %d, want 1", orderRepo.getAllCalls)
		}
		if tx.executed != 0 {
			t.Errorf("tx.Execute calls = %d, want 0 for read-only path", tx.executed)
		}
	})
}
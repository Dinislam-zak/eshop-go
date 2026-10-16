CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(15,4) NOT NULL
);

CREATE UNIQUE INDEX idx_product_name_active ON products(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_deleted_at ON products(deleted_at);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    number VARCHAR(255) NOT NULL,
    delivery_address TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_order_number_active ON orders(number) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);

CREATE TABLE order_item_models (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL
);

CREATE INDEX idx_order_item_models_deleted_at ON order_item_models(deleted_at);
CREATE INDEX idx_order_item_models_order_id ON order_item_models(order_id);
CREATE INDEX idx_order_item_models_product_id ON order_item_models(product_id);
CREATE UNIQUE INDEX idx_order_item_order_product ON order_item_models(order_id, product_id) WHERE deleted_at IS NULL;

ALTER TABLE order_item_models ADD CONSTRAINT fk_order_item_models_order FOREIGN KEY (order_id) REFERENCES orders(id);
ALTER TABLE order_item_models ADD CONSTRAINT fk_order_item_models_product FOREIGN KEY (product_id) REFERENCES products(id);
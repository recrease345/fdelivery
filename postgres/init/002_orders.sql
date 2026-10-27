CREATE TABLE IF NOT EXISTS orders.orders (
    id TEXT NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL,
    restaurant_id UUID NOT NULL,
    items JSONB NOT NULL,
    total_price BIGINT NOT NULL,
    delivery_address VARCHAR(300) NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT ck_orders_status CHECK(status IN('created', 'paid', 'preparing', 'ready', 'assigning', 'assigned', 'delivering', 'delivered', 'cancelled', 'failed_assign'))
);

CREATE INDEX idx_orders_user_id_created_at_desc ON orders.orders(user_id, created_at DESC);

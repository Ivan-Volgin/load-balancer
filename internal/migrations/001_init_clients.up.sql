CREATE TABLE clients (
                         client_id TEXT PRIMARY KEY,
                         capacity BIGINT NOT NULL,
                         rate_per_second BIGINT NOT NULL,
                         tokens BIGINT NOT NULL,
                         last_refill_at BIGINT NOT NULL
);

CREATE INDEX idx_client_id ON clients(client_id);
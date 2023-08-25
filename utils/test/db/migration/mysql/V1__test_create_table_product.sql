CREATE TABLE IF NOT EXISTS product(
    id VARCHAR(36) NOT NULL,
    name VARCHAR(255),
    code VARCHAR(100),

    CONSTRAINT pk_product_id PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE utf8_bin;

-- CREATE INDEX idx_prod_name ON product USING BTREE(name);
-- CREATE INDEX idx_prod_code ON product USING BTREE(code);

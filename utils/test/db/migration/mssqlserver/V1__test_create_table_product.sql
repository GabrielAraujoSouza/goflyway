IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE [TABLE_NAME] = 'product')
    CREATE TABLE product(
        id VARCHAR(36) NOT NULL,
        name VARCHAR(255),
        code VARCHAR(100),

        CONSTRAINT pk_product_id PRIMARY KEY (id)
    );

IF NOT EXISTS (SELECT * FROM SYS.INDEXES WHERE NAME='idx_prod_name')
    CREATE INDEX idx_prod_name ON product (name);

IF NOT EXISTS (SELECT * FROM SYS.INDEXES WHERE NAME='idx_prod_code')
    CREATE INDEX idx_prod_code ON product (code);

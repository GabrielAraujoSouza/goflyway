DO $$
BEGIN
	ALTER TABLE product
		ADD COLUMN description VARCHAR(255);
		
EXCEPTION
	WHEN duplicate_column THEN
		RAISE NOTICE 'Column already exists. Ignoring...';
	WHEN undefined_table THEN
		RAISE NOTICE 'Table "product" doesn''t exists. Ignoring...';
END$$;

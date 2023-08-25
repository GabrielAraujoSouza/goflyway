DO $$
BEGIN
	ALTER TABLE product
		DROP COLUMN description;
		
EXCEPTION
	WHEN undefined_column THEN
		RAISE NOTICE 'Column doesn''t exists. Ignoring...';
	WHEN undefined_table THEN
		RAISE NOTICE 'Table "product" doesn''t exists. Ignoring...';
END$$;

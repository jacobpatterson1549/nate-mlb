CREATE TABLE IF NOT EXISTS users
	( username VARCHAR(20) PRIMARY KEY
	, password TEXT
	);

-- TODO: add /admin/setup endpoint to create admin.
INSERT INTO users (username, password)
	SELECT 'admin', 'invalid_hash_value'
	WHERE NOT EXISTS (SELECT * FROM users WHERE username = 'admin')
	;

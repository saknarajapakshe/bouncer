CREATE TABLE IF NOT EXISTS permission_grants (
	subject_id TEXT NOT NULL,
	resource_id TEXT NOT NULL,
	permission_name TEXT NOT NULL,
	PRIMARY KEY (subject_id, resource_id, permission_name)
);

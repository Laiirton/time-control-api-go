CREATE TABLE IF NOT EXISTS time_records (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	event_type VARCHAR(100) NOT NULL,
	latitude NUMERIC(10,7) NOT NULL,
	longitude NUMERIC(10,7) NOT NULL,
	photo_path TEXT NOT NULL,
	photo_url TEXT NOT NULL,
	recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_time_records_user_recorded_at
	ON time_records (user_id, recorded_at DESC);

CREATE INDEX IF NOT EXISTS idx_time_records_recorded_at
	ON time_records (recorded_at DESC);
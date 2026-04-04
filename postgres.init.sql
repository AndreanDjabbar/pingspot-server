CREATE EXTENSION IF NOT EXISTS postgis;

-- after migration, run the following command to update the search_vector column for existing data
-- CREATE INDEX idx_reports_search_vector
-- ON reports USING GIN (search_vector);
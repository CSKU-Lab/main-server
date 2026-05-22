CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courses_name_trgm
  ON courses USING gin (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_labs_display_name_trgm
  ON labs USING gin (display_name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sections_name_trgm
  ON sections USING gin (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_materials_name_trgm
  ON materials USING gin (name gin_trgm_ops);

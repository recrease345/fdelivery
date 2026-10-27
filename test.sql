CREATE TYPE mood AS ENUM ('created');

SELECT pg_column_size('created'::text) AS text_bytes,
       pg_column_size('created'::mood) AS enum_bytes;

BEGIN;
ALTER TYPE mood ADD VALUE 'returning';
SELECT 'returning'::mood;
ROLLBACK;
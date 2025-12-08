-- Revert history table back to serial primary key
-- WARNING: This will regenerate IDs and lose UUID references

-- Step 1: Drop timestamp columns
ALTER TABLE history DROP COLUMN IF EXISTS updated_at;
ALTER TABLE history DROP COLUMN IF EXISTS created_at;

-- Step 2: Add temporary serial column
ALTER TABLE history ADD COLUMN old_id SERIAL;

-- Step 3: Drop primary key constraint
ALTER TABLE history DROP CONSTRAINT history_pkey;

-- Step 4: Drop UUID id column
ALTER TABLE history DROP COLUMN id;

-- Step 5: Rename old_id to id
ALTER TABLE history RENAME COLUMN old_id TO id;

-- Step 6: Add primary key constraint
ALTER TABLE history ADD PRIMARY KEY (id);

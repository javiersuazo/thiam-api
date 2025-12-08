-- Migrate history table from serial to UUID primary key
-- Step 1: Add new UUID column
ALTER TABLE history ADD COLUMN new_id UUID DEFAULT gen_random_uuid();

-- Step 2: Update all existing rows with UUIDs
UPDATE history SET new_id = gen_random_uuid() WHERE new_id IS NULL;

-- Step 3: Drop old primary key constraint
ALTER TABLE history DROP CONSTRAINT history_pkey;

-- Step 4: Drop old id column
ALTER TABLE history DROP COLUMN id;

-- Step 5: Rename new_id to id
ALTER TABLE history RENAME COLUMN new_id TO id;

-- Step 6: Add primary key constraint on new UUID column
ALTER TABLE history ADD PRIMARY KEY (id);

-- Step 7: Add created_at and updated_at for better tracking
ALTER TABLE history ADD COLUMN created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE history ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

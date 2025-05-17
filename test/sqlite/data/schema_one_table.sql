CREATE TABLE IF NOT EXISTS table_with_all_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    name TEXT NOT NULL,

    unique_number INTEGER UNIQUE,

    price REAL,

    is_active BOOLEAN DEFAULT TRUE,

    binary_data BLOB,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    age INTEGER CHECK (age >= 18),

    description TEXT DEFAULT 'No description'
);

CREATE INDEX IF NOT EXISTS idx_name ON table_with_all_types (name);

CREATE INDEX IF NOT EXISTS idx_unique_number_is_active ON table_with_all_types (unique_number, is_active);

CREATE TABLE IF NOT EXISTS table_with_all_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    name TEXT NOT NULL,

    unique_number INTEGER UNIQUE,
    number_with_default INTEGER DEFAULT 123,

    price REAL,
    price_with_default REAL DEFAULT 0.0,

    is_active BOOLEAN DEFAULT TRUE,

    binary_data BLOB,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    age INTEGER CHECK (age >= 18),

    description TEXT DEFAULT "Some description",

    empty_text TEXT DEFAULT "",
    single_char_text TEXT DEFAULT "x",
    single_quote_escape TEXT DEFAULT "It\'s a test",
    backslash_escape TEXT DEFAULT "C:\\Program Files",
    newline_escape TEXT DEFAULT "Line1\nLine2",
    tab_escape TEXT DEFAULT "Column1\tColumn2",
    unicode_escape TEXT DEFAULT "Unicode: \u263A"
);

CREATE INDEX IF NOT EXISTS idx_name ON table_with_all_types (name);

CREATE INDEX IF NOT EXISTS idx_unique_number_is_active ON table_with_all_types (unique_number, is_active);

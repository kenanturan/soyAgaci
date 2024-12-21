CREATE TABLE IF NOT EXISTS people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    identity_num TEXT,
    phone TEXT,
    birth_date TEXT,
    mother_id INTEGER,
    father_id INTEGER,
    gender TEXT CHECK (gender IN ('E', 'K')),
    about TEXT,
    photo_path TEXT,
    FOREIGN KEY (mother_id) REFERENCES people(id),
    FOREIGN KEY (father_id) REFERENCES people(id)
);

-- İndeksler
CREATE INDEX idx_names ON people(first_name, last_name);
CREATE INDEX idx_identity ON people(identity_num);
CREATE INDEX idx_parents ON people(mother_id, father_id); 
package main

import (
	"database/sql"
	"log"
)

// InitDB fonksiyonunun adını büyük harfle başlattık
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Basit bir tablo oluştur
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS people (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT,
		last_name TEXT,
		identity_num TEXT,
		phone TEXT,
		birth_date TEXT,
		mother_id INTEGER,
		father_id INTEGER,
		gender TEXT,
		about TEXT,
		photo_path TEXT
	)`)

	if err != nil {
		log.Printf("Tablo oluşturma hatası: %v", err)
		return nil, err
	}

	return db, nil
}

func GetPersonByID(db *sql.DB, id string) (*Person, error) {
	var person Person
	err := db.QueryRow(`
		SELECT id, first_name, last_name, identity_num, phone, birth_date, 
			mother_id, father_id, gender, about, photo_path
		FROM people WHERE id = ?`, id).Scan(
		&person.ID, &person.FirstName, &person.LastName, &person.IdentityNum,
		&person.Phone, &person.BirthDate, &person.MotherID, &person.FatherID,
		&person.Gender, &person.About, &person.PhotoPath)

	if err != nil {
		return nil, err
	}

	return &person, nil
}

func UpdatePerson(db *sql.DB, person Person) error {
	_, err := db.Exec(`
		UPDATE people 
		SET first_name = ?, last_name = ?, identity_num = ?, 
			phone = ?, birth_date = ?, gender = ?, 
			about = ?, photo_path = ?
		WHERE id = ?`,
		person.FirstName, person.LastName, person.IdentityNum,
		person.Phone, person.BirthDate, person.Gender,
		person.About, person.PhotoPath, person.ID)

	return err
}

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Person yapısı, veritabanındaki kişi bilgilerini temsil eder
type Person struct {
	ID          int
	FirstName   string
	LastName    string
	IdentityNum string // TC Kimlik No
	Phone       string // Telefon
	BirthDate   string
	MotherID    *int   // Anne ID'si (null olabilir)
	FatherID    *int   // Baba ID'si (null olabilir)
	Gender      string // "E" veya "K"
	About       string // Kişi hakkında bilgiler
	PhotoPath   string // Yeni alan
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	dbPath := "family_tree.db"

	// Veritabanı dosyası yoksa oluştur
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatal("Veritabanı dosyası oluşturulamadı:", err)
		}
		file.Close()
	}

	// Veritabanına bağlan
	var err error
	db, err = InitDB(dbPath)
	if err != nil {
		log.Fatal("Veritabanı başlatılamadı:", err)
	}
	defer db.Close()

	// Gin router'ı oluştur
	r := gin.Default()

	// Template fonksiyonlarını ekle
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
	})

	// Middleware ekleyelim
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Next()
	})

	// Hata yönetimi
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Bir hata oluştu",
				})
			}
		}()
		c.Next()
	})

	// HTML dosyalarını templates klasöründen yükle
	r.LoadHTMLGlob("templates/*")

	// Ana sayfa
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Kişi ekleme sayfası
	r.GET("/add", func(c *gin.Context) {
		c.HTML(http.StatusOK, "add.html", nil)
	})

	// Kişi ekleme işlemi
	r.POST("/add", func(c *gin.Context) {
		// Fotoğraf dosyasını al
		file, err := c.FormFile("photo")
		var photoPath string

		if err == nil && file != nil {
			// Dosya adını benzersiz yap
			filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
			photoPath = filepath.Join("uploads", filename)

			// uploads klasörünü oluştur
			if err := os.MkdirAll("uploads", 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Dosyayı kaydet
			if err := c.SaveUploadedFile(file, photoPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		person := Person{
			FirstName:   c.PostForm("firstName"),
			LastName:    c.PostForm("lastName"),
			IdentityNum: c.PostForm("identityNum"),
			Phone:       c.PostForm("phone"),
			BirthDate:   c.PostForm("birthDate"),
			Gender:      c.PostForm("gender"),
			About:       c.PostForm("about"),
			PhotoPath:   photoPath,
		}

		_, err = AddPerson(db, person)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	// Kişi arama
	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.HTML(http.StatusOK, "index.html", nil)
			return
		}

		people, err := SearchPerson(db, query, 1, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"People": people,
			"Query":  query,
		})
	})

	// Kişi düzenleme sayfası
	r.GET("/edit/:id", func(c *gin.Context) {
		id := c.Param("id")
		person, err := GetPersonByID(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"Person": person,
		})
	})

	// Kişi güncelleme işlemi
	r.POST("/edit/:id", func(c *gin.Context) {
		id := c.Param("id")
		log.Printf("Güncelleme isteği alındı. ID: %s", id)

		// Fotoğraf dosyasını al
		file, err := c.FormFile("photo")
		var photoPath string

		if err == nil && file != nil {
			// Dosya adını benzersiz yap
			filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
			photoPath = filepath.Join("uploads", filename)

			// uploads klasörünü oluştur
			if err := os.MkdirAll("uploads", 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Dosyayı kaydet
			if err := c.SaveUploadedFile(file, photoPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Mevcut kişiyi al
		person, err := GetPersonByID(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Yeni bilgileri güncelle
		person.FirstName = c.PostForm("firstName")
		person.LastName = c.PostForm("lastName")
		person.IdentityNum = c.PostForm("identityNum")
		person.Phone = c.PostForm("phone")
		person.BirthDate = c.PostForm("birthDate")
		person.Gender = c.PostForm("gender")
		person.About = c.PostForm("about")

		// Eğer yeni fotoğraf yüklendiyse güncelle
		if photoPath != "" {
			person.PhotoPath = photoPath
		}

		// Veritabanında güncelle
		err = UpdatePerson(db, *person)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	// Statik dosya sunucusu ekle
	r.Static("/uploads", "./uploads")

	log.Println("Web sunucusu http://localhost:8080 adresinde başlatılıyor...")
	r.Run(":8080")
}

func AddPerson(db *sql.DB, person Person) (int64, error) {
	// Temel validasyonlar
	if person.FirstName == "" || person.LastName == "" {
		return 0, fmt.Errorf("ad ve soyad boş olamaz")
	}

	if person.Gender != "E" && person.Gender != "K" {
		return 0, fmt.Errorf("geçersiz cinsiyet değeri")
	}

	// TC Kimlik kontrolü (11 haneli olmalı)
	if person.IdentityNum != "" && len(person.IdentityNum) != 11 {
		return 0, fmt.Errorf("TC Kimlik numarası 11 haneli olmalıdır")
	}

	result, err := db.Exec(`
        INSERT INTO people (first_name, last_name, identity_num, phone, birth_date, 
            mother_id, father_id, gender, about, photo_path)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		person.FirstName, person.LastName, person.IdentityNum, person.Phone,
		person.BirthDate, person.MotherID, person.FatherID, person.Gender,
		person.About, person.PhotoPath)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func SearchPerson(db *sql.DB, searchTerm string, page, limit int) ([]Person, error) {
	offset := (page - 1) * limit

	rows, err := db.Query(`
        SELECT id, first_name, last_name, identity_num, phone, birth_date, 
            mother_id, father_id, gender, about, photo_path
        FROM people
        WHERE first_name LIKE ? OR last_name LIKE ?
        LIMIT ? OFFSET ?`,
		"%"+searchTerm+"%", "%"+searchTerm+"%", limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		var p Person
		err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.IdentityNum,
			&p.Phone, &p.BirthDate, &p.MotherID, &p.FatherID, &p.Gender, &p.About, &p.PhotoPath)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}

	return people, nil
}

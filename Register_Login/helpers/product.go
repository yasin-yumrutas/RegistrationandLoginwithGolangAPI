package helpers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/sphinxql"
	_ "github.com/lib/pq"
)

// type user struct {
// 	ID       int
// 	Name     string
// 	Email    string
// 	Password string
// }

type User struct {
	ID              int    `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"-"`
	PasswordConfirm string `json:"--"`
	Email           string `json:"email"`
}

func Hey() {
	fmt.Println("say hey")
}

func Hata(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Product struct {
	Id                 int
	Title, Description string
	Price              float32
}

// const (
// 	host     = "localhost"
// 	port     = "5432"
// 	user     = "postgres"
// 	password = "321654"
// 	dbname   = "Postgresql"
// )

var db *sql.DB

func init() {
	var err error

	// connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ",
	//  host, port, user, password, dbname)

	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=321654 dbname=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(1 * time.Second)
	db.SetConnMaxLifetime(30 * time.Second)
	Hata(err)
}

func GetUserByUsername(username string) (*User, error) {
	// burada veritabanına bağlan ve ilgili kullanıcıyı bulmak için gerekli SQL sorgusunu yaz
	// daha sonra sorguyu çalıştır ve sonucu bir *sql.Rows nesnesinde depola
	rows, err := db.Query("SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// sonuçları döndürmeden önce tüm satırları döngüyle dolaşarak tek bir User nesnesiyle doldur
	user := &User{}
	for rows.Next() {
		err := rows.Scan(user.ID, user.Username, user.Email, user.PasswordConfirm)
		if err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// bulunan kullanıcıyı veya hata nesnesini döndür
	if user.ID == 0 {
		return nil, err
	}
	return user, nil
}

func InsertUser(u *User) error {
	_, err := db.Exec("INSERT INTO users(name, email, password) VALUES($1, $2, $3)", u.Username, u.Email, u.PasswordConfirm)
	if err != nil {
		return err
	}
	return nil
}

// func UpdateProduct(data Product) {
// 	result, err := db.Exec("UPDATE test SET title=$2 WHERE id=$1", data.Title, data.Id)
// 	Hata(err)
// 	rowsAffected, err := result.RowsAffected()
// 	Hata(err)
// 	fmt.Printf("Etkilenen Kayıt Sayısı:(%d)", rowsAffected)
// }

// func DeleteProduct(data Product) {
// 	result, err := db.Exec("DELETE FROM products WHERE id=2")
// 	Hata(err)
// 	rowsAffected, err := result.RowsAffected()
// 	Hata(err)
// 	fmt.Printf("Etkilenen Kayıt Sayısı:(%d)", rowsAffected)
// }

func GetAllUsers() ([]*User, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordConfirm)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func GetProductById(id int) {
	var product string
	err := db.QueryRow("SELECT title FROM test WHERE id=1", id).Scan(&product)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No product with that ID.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Product is %s\n", product)
	}
}

func SignUp(name, email, password string) error {
	// Check if email is already registered
	rows, err := db.Query("SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		return errors.New("email is already registered")
	}

	// Insert new user
	u := &User{
		Username:        name,
		Email:           email,
		PasswordConfirm: password,
	}
	if err := InsertUser(u); err != nil {
		return err
	}
	return nil
}

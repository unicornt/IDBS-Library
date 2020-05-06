package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	// mysql connector
	_ "github.com/go-sql-driver/mysql"

	sqlx "github.com/jmoiron/sqlx"
)

const (
	User     = ""
	Password = ""
	DBName   = "ass3"
)
var niltime time.Time

type Library struct {
	db *sqlx.DB
}

func (lib *Library) ConnectDB() {
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", User, Password, DBName))
	if err != nil {
		panic(err)
	}
	lib.db = db
}

// CreateTables created the tables in MySQL
func (lib *Library) CreateTables() error {
	_, err := lib.db.Exec(`DROP TABLE IF EXISTS Removedbook`);
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`DROP TABLE IF EXISTS Drecord`);
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`DROP TABLE IF EXISTS Record`);
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`DROP TABLE IF EXISTS Book`);
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`DROP TABLE IF EXISTS Student`);
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`CREATE TABLE Book(
						id INT UNSIGNED AUTO_INCREMENT,
			  			title VARCHAR(100),
			  			author VARCHAR(30),
			  			ISBN VARCHAR(13),
			  			PRIMARY KEY(id))`)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`CREATE TABLE Student(
						 id INT UNSIGNED AUTO_INCREMENT,
						 username VARCHAR(12) NOT NULL,
						 password VARCHAR(12),
						 PRIMARY KEY(id))`)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`CREATE TABLE Record(
						student_id INT UNSIGNED NOT NULL,
						book_id INT UNSIGNED NOT NULL,
						ddl DATE,
						ext_times INT,
						PRIMARY KEY(student_id, book_id)
					)`)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`CREATE TABLE Drecord(
					student_id INT UNSIGNED NOT NULL,
					book_id INT UNSIGNED NOT NULL,
					return_time DATE,
					PRIMARY KEY(student_id, book_id)
					)`)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`CREATE TABLE Removedbook(
			  			id INT UNSIGNED NOT NULL,
			  			title VARCHAR(100),
			  			author VARCHAR(30),
			  			ISBN VARCHAR(13),
			  			removereason VARCHAR(100),
			  			PRIMARY KEY(id)
			  )`)
	return err
}
type Book struct {
	book_id int
	title string
	author string
	ISBN string
	ddl time.Time
	ext_times int
}

// AddBook add a book into the library
func (lib *Library) AddBook(title, author, ISBN string) error {
	_, err := lib.db.Exec(`INSERT INTO Book(title, author, ISBN) VALUES (?, ?, ?)`, title, author, ISBN)
	return err
}

// remove a book with reason
func (lib *Library) RemoveBook(book_id int, reason string) error {
	result, err := lib.db.Query(`SELECT *
								 FROM Book
								 WHERE id = ?`, book_id)
	defer func(){
		err = result.Close()
	}()
	if err != nil {
		return err
	}
	if !result.Next() {
		err = errors.New("error: No such book")
		return err
	}
	_, err = lib.db.Exec(`DELETE FROM Book
						   WHERE id = ?`, book_id)
	if err != nil {
		return err
	}
	var title, author, ISBN string
	err = result.Scan(&book_id, &title, &author, &ISBN)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`INSERT INTO Removedbook(id, title, author, ISBN, removereason) VALUES (?, ?, ?, ?, ?)`, book_id, title, author, ISBN, reason)
	return err
}

func (lib* Library) AddStudent(username, password string) error {
	result, err := lib.db.Query(`SELECT *
						 		 FROM Student
						 		 WHERE username = ?`, username)
	defer func(){
		result.Close()
	}()
	if result.Next() {
		err = errors.New("username already exists")
		return err
	}
	_, err = lib.db.Exec(`INSERT INTO Student(username, password) VALUES (?, ?)`, username, password)
	return err
}

// input a string, 0 for either title, author, ISBN = input  1, 2, 3 for certain item
func (lib *Library) QueryBook(input string, mode int) ([]Book, error) {
	var item [4]string
	item[1] = "title"
	item[2] = "author"
	item[3] = "ISBN"
	var len int
	var result *sql.Rows
	var err error
	defer func(){
		err = result.Close()
	}()
	if mode == 0 {
		result, err = lib.db.Query(`SELECT id, title, author, ISBN
									FROM Book
									WHERE title = ? OR author = ? OR ISBN = ?`, input, input, input)
		if err != nil {
			return nil, err
		}
		err = lib.db.QueryRow(`SELECT COUNT(*)
							   FROM Book
							   WHERE title = ? OR author = ? OR ISBN = ?`, input, input, input).Scan(&len)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = lib.db.Query(fmt.Sprintf(`SELECT id, title, author, ISBN
									   FROM Book
									   WHERE %s = ?`, item[mode]), input)
		if err != nil {
			return nil, err
		}
		err = lib.db.QueryRow(fmt.Sprintf(`SELECT COUNT(*)
									   FROM Book
									   WHERE %s = ?`, item[mode]), input).Scan(&len)
		if err != nil {
			return nil, err
		}
	}
	ret := make([]Book, len)
	var book_id int
	var title, author, ISBN string
	i := 0
	for result.Next() {
		err = result.Scan(&book_id, &title, &author, &ISBN)
		if err != nil {
			return nil, err
		}
		ret[i] = Book{book_id, title, author, ISBN, niltime, -1}
		i = i + 1
	}
	return ret, nil
}
//borrow a book with student account, input student id and book id
func (lib *Library) BorrowBook(student_id, book_id int, brtime time.Time) error {
	today := time.Now()
	result, err := lib.db.Query(`SELECT *
								FROM Book
								WHERE id = ?`, book_id)
	if err != nil {
		return err
	}
	if !result.Next() {
		err = errors.New("Book not exists")
		return err
	}
	result, err = lib.db.Query(`SELECT *
								FROM Record
								WHERE book_id = ?`, book_id)
	if err != nil {
		return err
	}
	var overdue_times int
	if result.Next() {
		err = errors.New("Book has been borrowed")
		return err
	}
	d, _ := time.ParseDuration("24h")
	ddl := brtime.Add(28 * d)
	err = lib.db.QueryRow(`SELECT COUNT(*)
						   FROM Record
						   WHERE student_id = ? AND ddl < ?`, student_id, today).Scan(&overdue_times)
//	fmt.Println(overdue_times)
	if err != nil {
		return err
	}
	if overdue_times > 3 {
		err = errors.New("Student account suspended")
		return err
	}
//	fmt.Println(student_id, book_id, ddl)
	_, err = lib.db.Exec(`INSERT INTO Record(student_id, book_id, ddl, ext_times) VALUES (?, ?, ?, 0)`, student_id, book_id, ddl)
	return err
}

//query a student's borrow history, first is the book that hasn't returned second is the book that has returned
func (lib *Library) QueryHistory(student_id int) ([]Book, []Book, error) {
	result, err := lib.db.Query(`SELECT id, title, author, ISBN, ddl, ext_times
								 FROM Record, Book
								 WHERE student_id = ? AND Book.id = Record.book_id`, student_id)
	defer func(){
		result.Close()
	}()
	if err != nil {
		return nil,nil,err
	}
	var book_id, ext_times int
	var ddl time.Time
	var title, author, ISBN string
	var len1, len2 int
	err = lib.db.QueryRow(`SELECT COUNT(*)
						   FROM Record, Book
						   WHERE student_id = ? AND Book.id = Record.book_id`, student_id).Scan(&len1)
	if err != nil {
		return nil,nil,err
	}
	ret1 := make([]Book, len1)
	i := 0
	for result.Next() {
		err = result.Scan(&book_id, &title, &author, &ISBN, &ddl, &ext_times)
		if err != nil {
			return nil, nil, err
		}
		ret1[i] = Book{book_id, title, author, ISBN, ddl, ext_times}
		i = i + 1
	}
	result, err = lib.db.Query(`SELECT book_id, title, author, ISBN, return_time
								 FROM Drecord, Book
								 WHERE student_id = ? AND Book.id = Drecord.book_id`, student_id)
	if err != nil {
		return nil,nil,err
	}
	err = lib.db.QueryRow(`SELECT COUNT(*)
						 FROM Drecord, Book
						 WHERE student_id = ? AND Book.id = Drecord.book_id`, student_id).Scan(&len2)
	if err != nil {
		return nil,nil,err
	}
	ret2 := make([]Book, len2)
	i = 0
	for result.Next() {
		err = result.Scan(&book_id, &title, &author, &ISBN, &ddl)
		if err != nil {
			return nil, nil, err
		}
		ret2[i] = Book{book_id, title, author, ISBN, ddl, -1}
		i = i + 1
	}
	//fmt.Println(len1, len2)
	return ret1, ret2, nil
}

//query a student's not returned book
func (lib *Library) QueryNotReturn(student_id int) ([]Book, error) {
	result, err := lib.db.Query(`SELECT book_id, title, author, ISBN, ddl, ext_times
								 FROM Record, Book
								 WHERE student_id = ? AND Book.id = Record.book_id`, student_id)
	var len int
	defer func (){
		result.Close()
	}()
	err = lib.db.QueryRow(`SELECT COUNT(*)
				  		   FROM Record, Book
				 		   WHERE student_id = ? AND Book.id = Record.book_id`, student_id).Scan(&len)
	var book_id, ext_times int
	var ddl time.Time
	var title, author, ISBN string
	ret := make([]Book, len)
	i := 0
	for result.Next() {
		err = result.Scan(&book_id, &title, &author, &ISBN, &ddl, &ext_times)
		if err != nil {
			return nil, err
		}
		ret[i] = Book{book_id, title, author, ISBN, ddl, ext_times}
		i = i + 1
	}
	return ret, nil
}

//check deadline of a returning book
func (lib *Library) QueryDeadline(book_id int) (time.Time, error) {
	result, err := lib.db.Query(`SELECT ddl
								 FROM Record
								 WHERE book_id = ?`, book_id)
	var ddl time.Time
	if result.Next() {
		err = result.Scan(&ddl)
		if err != nil {
			return niltime, err
		}
		return ddl, nil
	} else {
		err = errors.New("This book hasn't been brrowed or not exists")
		return niltime, err
	}
	return niltime, nil
}

//extend deadline of a book
func (lib *Library) ExtendTime(book_id int) error {
	result, err := lib.db.Query(`SELECT ddl
								 FROM Record
								 WHERE book_id = ? AND ext_times < 3`, book_id)
	if err != nil {
		return err
	}
	if result.Next() {
		lib.db.Exec(`UPDATE Record SET ddl = ddl + 7, ext_times = ext_times + 1 WHERE book_id = ?`, book_id)
	} else {
		err = errors.New("Extend operation refused")
		return err
	}
	return nil
}

//check overdue books
func (lib *Library) QueryOverdue(student_id int, today time.Time) ([]Book, error) {
	year, month, day := today.Date()
	result, err := lib.db.Query(fmt.Sprintf(`SELECT book_id, title, author, ISBN, ddl
								 FROM Record, Book
								 WHERE ddl < '%d-%02d-%02d' AND student_id = %d`, year, month, day, student_id))
	if err != nil {
		return nil, err
	}
	var len int
	err = lib.db.QueryRow(fmt.Sprintf(`SELECT COUNT(*)
						 FROM Record
						 WHERE ddl < '%d-%02d-%02d' AND student_id = %d`, year, month, day, student_id)).Scan(&len)
	//fmt.Println(len)
	if err != nil {
		return nil, err
	}
	var book_id, i int
	var title, author, ISBN string
	var ddl time.Time
	ret := make([]Book, len)
	for result.Next() {
		err = result.Scan(&book_id, &title, &author, &ISBN, &ddl)
		if err != nil {
			return nil, err
		}
		ret[i] = Book{book_id, title, author, ISBN, ddl, -1}
		i = i + 1
	}
	return ret, nil
}

//return book
func (lib *Library) ReturnBook(student_id, book_id int) error {
	today := time.Now()
	result, err := lib.db.Query(`SELECT *
								 FROM Record
								 WHERE student_id = ? AND book_id = ?`, student_id, book_id)
	if !result.Next() {
		err = errors.New("Do not have this record!")
		return err
	}
	_, err = lib.db.Exec(`INSERT INTO Drecord (student_id, book_id, return_time) VALUES (?, ?, ?)`, student_id, book_id, today)
	if err != nil {
		return err
	}
	_, err = lib.db.Exec(`DELETE FROM Record WHERE student_id = ? AND book_id = ?`, student_id, book_id)
	return err
}
func main() {
	fmt.Println("Welcome to the Library Management System!")
	/*fmt.Println("Please enter your username: ")
	var username, password string
	fmt.Scanln(&username)
	fmt.Println("Please enter your password:")
	fmt.Scanln(&password)
	/*
		check username and password module
	*/
	/*if username == "admin" {
		//logic as admin
		fmt.Println("Input:\n0 to add a book\n1 to remove a book\n2 to add a student account\n3 to query a book\n4 ")
	else {

	}*/
}
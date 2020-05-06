package main

import (
	"testing"
	"fmt"
	"reflect"
	"time"
)
var now time.Time
func TestInit(t *testing.T) {
	d := time.Now().UTC()
	now = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}
func TestCreateTables(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	err := lib.CreateTables()
	if err != nil {
		t.Errorf("can't create tables")
	}
}

func TestAddBook(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	for i := 0 ; i < 3; i++ {
		err := lib.AddBook("CS:APP", "Bryant", "9787111561279")
		if err != nil {
			t.Errorf("can't add book 1")
		}
	}
	err := lib.AddBook("ITA", "Cormen", "9787111407010")
	if err != nil {
		t.Errorf("can't add book 2")
	}
	err = lib.AddBook("alice", "bob", "9787111407012")
	if err != nil {
		t.Errorf("can't add book 3")
	}
	err = lib.AddBook("bob", "alice", "9787111407013")
	if err != nil {
		t.Errorf("can't add book 2")
	}
}

func TestRemoveBook(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	err := lib.RemoveBook(1, "lost")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestAddStudent(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	err := lib.AddStudent("unicornt", "123")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestQueryBook(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	var book [2]Book
	book[0] = Book{5, "alice", "bob", "9787111407012", niltime, -1}
	book[1] = Book{6, "bob", "alice", "9787111407013", niltime, -1}
	var tests = []struct {
		input string
		mode int
		want []Book
	}{
		{"alice", 0, []Book{book[0], book[1]}},
		{"alice", 1, []Book{book[0]}},
		{"alice", 2, []Book{book[1]}},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s, %d", tt.input, tt.mode)
		t.Run(testname, func(t *testing.T) {
			var ans, err = lib.QueryBook(tt.input, tt.mode)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("Query result not meet the ans")
			}
		})
	}
}

func TestBorrowBook(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	d, _ :=time.ParseDuration("24h")
	brtime2 := now.Add(-30 * d)
	brtime3 := now.Add(-29 * d)
	brtime4 := now.Add(-31 * d)
	brtime5 := now.Add(-32 * d)
	var tests = []struct {
		sid, bid int
		brtime time.Time
		wanterr bool
	}{
		{1, 2, brtime2, false},
		{1, 3, brtime3, false},
		{1, 4, brtime4, false},
		{1, 5, brtime5, false},
		{1, 6, brtime5, true},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%d %d", tt.sid, tt.bid)
		t.Run(testname, func(t *testing.T) {
			err := lib.BorrowBook(tt.sid, tt.bid, tt.brtime)
			if err != nil && tt.wanterr == false {
				t.Errorf(err.Error())
			}
			if err == nil && tt.wanterr == true {
				t.Errorf("Shoudn't be allowed to borrow!")
			}
		})
	}
}

func TestReturnBook(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	err := lib.ReturnBook(1, 5)
	if err != nil {
		t.Errorf("can't return book")
	}
}

func TestQueryHistory(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	ret1, ret2, err := lib.QueryHistory(1)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !(ret1[0].book_id == 2 && ret1[1].book_id == 3 && ret1[2].book_id == 4 && ret2[0].book_id == 5) {
		t.Errorf("History is wrong")
	}
}

func TestQueryNotReturn(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	ret, err := lib.QueryNotReturn(1)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !(ret[0].book_id == 2 && ret[1].book_id == 3 && ret[2].book_id == 4 ) {
		t.Errorf("query not return book error")
	}
}

func TestQueryDeadline(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	ret, err := lib.QueryDeadline(2)
	if err != nil {
		t.Errorf(err.Error())
	}
	d, _ := time.ParseDuration("24h")
	if ret != now.Add(-2 * d) {
		t.Errorf("ddl query error")
	}
	//fmt.Println(ret, now)
}

func TestExtendTime(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	err := lib.ExtendTime(2)
	if err != nil {
		t.Errorf(err.Error())
	}
	ret, err := lib.QueryDeadline(2)
	d, _ := time.ParseDuration("24h")
	if ret != now.Add(5 * d) {
		t.Errorf("ddl extend error")
	}
	err = lib.ExtendTime(2)
	err = lib.ExtendTime(2)
	err = lib.ExtendTime(2)
	if err == nil {
		t.Errorf("shoudn't be allowed to extend")
	}
}

func TestQueryOverdue(t *testing.T) {
	lib := Library{}
	lib.ConnectDB()
	ret, err := lib.QueryOverdue(2, now)
	if err != nil {
		t.Errorf(err.Error())
	}
	for _, tt := range ret {
		fmt.Println(reflect.TypeOf(tt))
		if tt.book_id == 1 || tt.book_id == 2 || tt.book_id > 5 {
			t.Errorf("Checkoverdue error")
		}
	}
}
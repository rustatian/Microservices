package dboperation

import "database/sql"
import (
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"Models"
)


func UserRegistration(userModel Models.User) {


	db, err := sql.Open("mysql", "root:ZXCfdsa1208@tcp(18.195.2.253:3306)/TaskCalendarDb")
	if err != nil {
		panic(err.Error())
	}

	//db.Query()
	ret, err := db.Exec("CREATE TABLE User (ID INT NOT NULL, Username TEXT(100) NOT NULL , FullName TEXT(100) NOT NULL, email TEXT(150) NOT NULL, PasswordHash TEXT(500) NOT NULL, PasswordSalt TEXT(500) NULL, IsDisabled BOOL NOT NULL, PRIMARY KEY (ID), UNIQUE INDEX ID_UNIQUE (ID ASC));")
	if err != nil{
		panic(err)
	}

	fmt.Println(ret)
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
}


package dboperation

import "database/sql"
import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"models"
	"os"
)

var createTable = "CREATE TABLE User (ID INT NOT NULL, Username TEXT(100) NOT NULL, FullName TEXT(100) NOT NULL, " +
	"email TEXT(150) NOT NULL, PasswordHash TEXT(500) NOT NULL, jwtToken TEXT(500) NOT NULL, IsDisabled BOOL NOT NULL, " +
	"PRIMARY KEY (ID), UNIQUE INDEX ID_UNIQUE (ID ASC));"

func configuration(configName string) models.Configuration {
	file, _ := os.Open(configName)
	defer file.Close()
	decode := json.NewDecoder(file)
	configuration := models.Configuration{}
	err := decode.Decode(&configuration)
	if err != nil {
		panic(err.Error())
	}
	return configuration
}

func CreateDatabase() bool {
	var config = configuration("config.json")

	db, err := sql.Open("mysql", config.DbCreds)
	if err != nil {
		panic(err.Error())
		return false
	}

	ret, err := db.Exec(createTable)
	if err != nil {
		panic(ret)
		return false
	}

	return true
}

func WriteDataToDb(user models.User) bool {
	var config = configuration("config.json")
	db, err := sql.Open("mysql", config.DbCreds)
	if err != nil {
		panic(err.Error())
		return false
	}
	defer db.Close()

	stmIns, err := db.Prepare("INSERT INTO User (Username, FullName, email, PasswordHash, jwtToken, IsDisabled) VALUES (?, ?, ?, ?, ?, ?);")
	defer stmIns.Close()

	_, err = stmIns.Exec(user.Username, user.FullName, user.Email, user.PasswordHash, user.JsonToken, user.IsDisables)
	if err != nil {
		panic(err.Error())
		return false
	}

	return true
}

func GetHashFromDb(user models.User) (string, bool) {
	var config = configuration("config.json")

	db, err := sql.Open("mysql", config.DbCreds)
	if err != nil {
		panic(err.Error())
		return "", false
	}
	defer db.Close()

	sel, err := db.Prepare("SELECT PasswordHash FROM User WHERE Username = ?;")
	if err != nil {
		panic(err.Error())
		return "", false
	}
	defer sel.Close()

	var hash string
	err = sel.QueryRow(user.Username).Scan(&hash)
	if err != nil {
		panic(err.Error())
		return "", false
	}

	return hash, true
}

func UpdateTokenForUser(user models.User) bool {
	var config = configuration("config.json")
	db, err := sql.Open("mysql", config.DbCreds)

	if err != nil {
		panic(err.Error())
		return false
	}
	defer db.Close()

	upd, err := db.Prepare("UPDATE User SET jwtToken = ? WHERE Username = ?;")
	if err != nil {
		panic(err.Error())
		return false
	}

	defer upd.Close()

	_, err = upd.Exec(user.JsonToken, user.Username)
	if err != nil {
		panic(err.Error())
		return false
	}

	return true
}

func CheckifUserExist(user string) bool {
	var config = configuration("config.json")
	db, err := sql.Open("mysql", config.DbCreds)
	if err != nil {
		panic(err.Error())
		return false
	}
	defer db.Close()

	sel, err := db.Prepare("SELECT ID FROM User WHERE Username = ?;")
	if err != nil {
		panic(err.Error())
		return false
	}
	defer sel.Close()

	var id int

	err = sel.QueryRow(user).Scan(&id)
	if err != nil { //NoRows error - is good, user does no exist
		return false
	} else {
		return true // else - user exist
	}

}

func CheckifMailExist(user models.User) bool {
	var config = configuration("config.json")
	db, err := sql.Open("mysql", config.DbCreds)
	if err != nil {
		panic(err.Error())
		return false
	}
	defer db.Close()

	sel, err := db.Prepare("SELECT ID FROM User WHERE email = ?;")
	if err != nil {
		panic(err.Error())
		return false
	}
	defer sel.Close()

	var id int

	err = sel.QueryRow(user.Email).Scan(&id)
	if err != nil { //NoRows error - is good, user does no exist
		return false
	} else {
		return true // else - user exist
	}
}

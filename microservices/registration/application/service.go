package application

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type User struct {
	Username   string
	FullName   string
	Email      string
	Password   string
	IsDisabled bool
}

var (
	dbCreds      string
	consAddr     string
	vaultSvcName string
	svcTag       string
)

func init() {
	if dev := os.Getenv("DEV"); dev == "False" {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverProd")
		consAddr = viper.GetString("Consul.addressProd")
		vaultSvcName = viper.GetString("services.vault") //to get hash from pass
		svcTag = viper.GetString("tags.tag")
	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverDev")
		consAddr = viper.GetString("Consul.addressDev")
		vaultSvcName = viper.GetString("services.vault") //to get hash from pass
		svcTag = viper.GetString("tags.tag")
	}

}

type DB struct {
	*sqlx.DB
}

func NewPostgresDb(url string) *DB {
	db := sqlx.MustOpen("postgres", url)
	return &DB{db}
}

func (db *DB) Registration(user *User) (bool, error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var hresp hashResponse
	//TODO create struct
	var req []byte = []byte(`{"password":"` + password + `"}`)

	addr, err := svcdiscovery.ServiceDiscovery().Find(&consAddr, &vaultSvcName, &svcTag)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(addr+"/hash", "application/json; charset=utf-8", bytes.NewBuffer(req))
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &hresp)

	if hresp.Err != "" || err != nil {
		return false, err
	}

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}

	stmIns, err := db.Query(`INSERT INTO "User" (username, fullname, email, passwordhash, isdisabled) VALUES ($1, $2, $3, $4, $5);`, username, fullname, email, hresp.Hash, false)
	if err != nil {
		return false, err.(*pq.Error)
	}
	defer stmIns.Close()

	return true, nil
}

func (db *DB) UsernameValidation(user *User) (bool, error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}
	sel, err := db.Prepare(`SELECT id FROM "User" WHERE username = $1;`)
	if err != nil {
		return false, err.(*pq.Error)
	}
	defer sel.Close()

	var id int
	err = sel.QueryRow(username).Scan(&id)
	if err != nil { //NoRows error - user does no exist
		return false, nil
	} else {
		return true, nil // else - user exist
	}
}

func (db *DB) EmailValidation(user *User) (bool, error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}

	sel, err := db.Prepare(`SELECT id FROM "User" WHERE email = $1;`)
	if err != nil {
		return false, err.(*pq.Error)
	}
	defer sel.Close()

	var id int

	err = sel.QueryRow(email).Scan(&id)
	if err != nil { //NoRows error - email does no exist
		return false, nil
	} else {
		return true, nil // else - email exist
	}
}

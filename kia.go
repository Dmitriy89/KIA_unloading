package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"golang.org/x/text/encoding/charmap"
)

var (
	pass    string
	alldata bool
	db      *sql.DB
)

func init() {
	flag.StringVar(&pass, "p", "MyPassword", "Password")
	flag.BoolVar(&alldata, "a", false, "Not only yesterday")
	flag.Parse()
}

func request() error {
	var (
		read      [][]string
		tx        *sql.Tx
		stmt      *sql.Stmt
		resp      *http.Response
		yesterday string
		err       error
	)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(time.Minute),
	}

	if alldata {
		yesterday = "N"
	} else {
		yesterday = "Y"
	}

	form := url.Values{
		"password":  {pass},
		"yesterday": {yesterday},
	}

	params := bytes.NewBufferString(form.Encode())
	resp, err = client.Post("http://riox-line.kia.ru/output101/out.php", "application/x-www-form-urlencoded", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(errors.New("Password is wrong?"))
	}

	decodeBin := charmap.Windows1251.NewDecoder().Reader(resp.Body)
	newReadCSV := csv.NewReader(decodeBin)
	newReadCSV.Comma = ';'
	newReadCSV.LazyQuotes = true

	read, err = newReadCSV.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	genSlice := make([][]string, len(read)-1)
	copy(genSlice, read[1:])

	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(`INSERT INTO [EVENT].[Car display_Kia Rio X-Line_Dec_2017](
									  [City]
									  ,[Dealer]
									  ,[Lastname]
									  ,[Firstname]
									  ,[Email]
									  ,[Mobilephone]
									  ,[CreateDate]) VALUES(?,?,?,?,?,?,?);`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, val := range genSlice {
		_, err = stmt.Exec(val[0], val[1], val[2], val[3], val[4], val[5], val[6])
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func main() {
	var err error
	db, err = sql.Open("mssql", "server=myServer;user id=myID;password=myPass;database=myDB;connection timeout=200")
	if err != nil {
		log.Fatal(err)
		fmt.Println("Не подключился к БД")
	} else {
		fmt.Println("Подключился к БД")
	}

	err = request()
	if err != nil {
		log.Fatal(err)
	}

	/*rows, err := db.Query("SELECT [City] FROM [EVENT].[Car display_Kia Rio X-Line_Dec_2017]")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var city string
		rows.Scan(&city)
		fmt.Println(city)
	}*/
	/*if err != nil {
		log.Println("Ошибка:",err)
		os.Exit(2)
	}*/
}

package main

import (
	"fmt"
	"log"
	"database/sql"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var stmt *sql.Stmt
var err error

func InitDB(dataSourceName string) {
    db, err = sql.Open("sqlite3", dataSourceName)
    if err != nil {
        log.Panic(err)
    }
	stmt, err = db.Prepare(`select * from urls where url = ?`)
}

func UrlInfo(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	var url string
	var threat string
	var dateadded string
	row := stmt.QueryRow(ctx.UserValue("url"))
	switch err = row.Scan(&url, &threat, &dateadded); err {
	case sql.ErrNoRows:
		fmt.Fprintf(ctx, "{\"results\": [{}]}")
	case nil:
		s := []byte("{\"results\": [{\"url\": \"")
        s = append(s, url...)
        s = append(s, "\", \"threat\": \""...)
        s = append(s, threat...)
        s = append(s, "\", \"dateadded\": \""...)
        s = append(s, dateadded...)
        s = append(s, "\"}]}"...)
        fmt.Fprintf(ctx, string(s))
	default:
		panic(err)
	}
}

func main() {
	InitDB("~/rqlited/db.sqlite?cache=shared&mode=r")
	r := router.New()
	r.GET("/urlinfo/1/*url", UrlInfo)
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
	stmt.Close()
	db.Close()
}

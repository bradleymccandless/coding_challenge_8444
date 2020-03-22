package main

import (
	"fmt"
	"log"
	"database/sql"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	_ "github.com/mattn/go-sqlite3"
)

func UrlInfo(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	db, err := sql.Open("sqlite3", "~/rqlited/db.sqlite")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
	sqlStatement := `select * from urls where url = $url`
	var url string
	var threat string
	var dateadded string
	row := db.QueryRow(sqlStatement, ctx.UserValue("url"))
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
	r := router.New()
	r.GET("/urlinfo/1/*url", UrlInfo)
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}

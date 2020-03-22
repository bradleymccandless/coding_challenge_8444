package main

import (
	"io"
	"os"
	"fmt"
	"log"
	"strings"
	//"net/url"
	"net/http"
	"io/ioutil"
	"encoding/csv"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)
func rqliteApi(sql []byte, action string) []byte {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fmt.Sprintf("http://localhost:4001/db/%s", action))
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(sql)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	client.Do(req, resp)
	bodyBytes := resp.Body()
	return bodyBytes
}
func getUrlHaus() (string, error) {
	// grab a fresh URL list
	// this thing can grow bigger than FreeRam so we cant use fasthttp
	resp, err := http.Get("https://urlhaus.abuse.ch/downloads/csv/")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    out, err := ioutil.TempFile("/tmp", "*.csv")
    if err != nil {
        return "", err
    }
    defer out.Close()
    _, err = io.Copy(out, resp.Body)
    return out.Name(), err
}
func importUrls(urlHausCsv string) {
	f, err := os.Open(urlHausCsv)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 8
	r.Comment = '#'
	lineCount := 0
	// 100000 chunk size: no sweat. 16s 400MB RAM
	// but we only need 5000 every 10mins :)
	// which will take about 0.8s and 20MB RAM
	chunkSize := 100000
	transaction := []byte("[")
	var resp []byte
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}
		// we only need http proto
		if strings.HasPrefix(record[2], "http://"){
			if lineCount == 0 {
				transaction = append(transaction, "\n\t"...)
			} else {
				transaction = append(transaction, ",\n\t"...)
			}
			transaction = append(transaction, "\"insert into urls ("...)
			transaction = append(transaction, "dateadded, url, threat) values("...)
			for i, column := range record {
				if i == 1 { 
					transaction = append(transaction, "'"...)
	 	    		transaction = append(transaction, column...)
   		    		transaction = append(transaction, "'"...)
				}
				if i == 2 {
					transaction = append(transaction, ", '"...)
					// massage the data...
					if strings.Contains(string(strings.Split(record[2], "/")[2]), ":") {
						transaction = append(transaction, strings.TrimPrefix(record[2], "http://")...)
					} else { 
						transaction = append(transaction, strings.Split(record[2], "/")[2]...)
						transaction = append(transaction, ":80/"...)
						transaction = append(transaction, strings.Join(strings.Split(record[2], "/")[3:], "/")...)
					}
   	       	        transaction = append(transaction, "'"...)
				}
				if i == 4 {
   	       		   	transaction = append(transaction, ", '"...)
   	       	        transaction = append(transaction, column...)
   	       	        transaction = append(transaction, "'"...)
   	      		}
			}
			lineCount += 1
			transaction = append(transaction, ")\""...)
			if lineCount == chunkSize {
				transaction = append(transaction, "\n]"...)
				resp = rqliteApi(transaction, "execute&transaction")
				if fastjson.GetString(resp, "results", "0", "error") != "" {
   	    			log.Println(fastjson.GetString(resp, "results", "0", "error"))
				}
				lineCount = 0
				transaction = []byte("[")
				//update a new chunk of URLs every 10 minutes
				//time.Sleep(10 * time.Minute)
			}
		}
	}
	// leftover rows smaller than a whole chunk
	if len(string(transaction)) > 1 {
		transaction = append(transaction, "\n]"...)
		resp = rqliteApi(transaction, "execute&transaction")
        if fastjson.GetString(resp, "results", "0", "error") != "" {
        	log.Fatal(fastjson.GetString(resp, "results", "0", "error"))
        }
	}
	return
}
func main() {
	resp := rqliteApi([]byte("[\"PRAGMA journal_mode=WAL\"]"), "execute")
    if fastjson.GetString(resp, "results") != "" {
        log.Fatal(fastjson.GetString(resp, "results"))
    }
	resp = rqliteApi([]byte(`["drop table urls;"]`), "execute")
	if fastjson.GetString(resp, "results", "0", "error") != "" {
		log.Println(fastjson.GetString(resp, "results", "0", "error"))
	}
	resp = rqliteApi([]byte("[\"CREATE TABLE urls(url text primary key, threat text, dateadded datetime)\"]"), "execute")
	if fastjson.GetString(resp, "results") != "" {
		log.Fatal(fastjson.GetString(resp, "results"))
	}
	urlHausCsv, err := getUrlHaus()
    if err != nil {
       	log.Fatal(err)
    }
	importUrls(urlHausCsv)
	os.Remove(urlHausCsv)
	log.Println("Done")
}

package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var articles []article

type Data struct {
	Article  *article
	Articles *[]article
	Tags string
}

type article struct {
	Id                     int
	Title, Body, Url string
	Tags []string
	Created_at, Updated_at time.Time
}

func createArticle(title, body, url string, tags []string, created_at, updated_at time.Time) article {
	return article{
		Title:      title,
		Body:       body,
		Tags:       tags,
		Url:        url,
		Created_at: created_at.Local(),
		Updated_at: updated_at.Local(),
	}
}

func (a *article) save() error {
	if err := a.insert(); err != nil {
		if err := a.update(); err != nil {
			return err
		}
	}
	return nil
}

func (a *article) insert() error {
	_, err := db.Exec("INSERT INTO articles (title, body, tags, url) VALUES (?, ?, ?, ?)", a.Title, a.Body, strings.Join(a.Tags, ","), a.Url)
	if err != nil {
		return err
	}
	return nil
}

func (a *article) update() error {
	_, err := db.Exec("UPDATE articles SET title = ?, body = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE url = ?", a.Title, a.Body, strings.Join(a.Tags, ","), a.Url)
	if err != nil {
		return err
	}
	return nil
}

func findArticleByUrl(url string) (article, error) {
	row := db.QueryRow("SELECT * FROM articles WHERE url = ?", url)
	a := article{}
	var tags string
	if err := row.Scan(&a.Id, &a.Title, &a.Body, &tags, &a.Url, &a.Created_at, &a.Updated_at); err != nil {
		return article{}, err
	}
	a.Tags = strings.Split(tags, ",")
	return a, nil
}

func findArticlesByTag(tag string) ([]article, error) {
	var articles []article
	var a article

	m := []string{"%," + tag + ",%", tag + ",%", "%," + tag, tag}
	rows, err := db.Query("SELECT * FROM articles WHERE tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?", m[0], m[1], m[2], m[3])
	if err != nil {
		return []article{}, err
	}
	var tags string
	for rows.Next() {
		if err := rows.Scan(&a.Id, &a.Title, &a.Body, &tags, &a.Url, &a.Created_at, &a.Updated_at); err != nil {
			return []article{}, err
		}
		a.Tags = strings.Split(tags, ",")
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return []article{}, err
	}
	return articles, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	d := Data{Article: &a, Articles: &articles}
	t, _ := template.ParseFiles("view.html")
	t.Execute(w, &d)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/edit/"):]
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	t, _ := template.ParseFiles("edit.html")
	d := Data{Article: &a, Tags: strings.Join(a.Tags, ",")}
	t.Execute(w, &d)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/save/"):]

	a := article{
		Title: r.FormValue("title"),
		Body: r.FormValue("body"),
		Url: url,
		Tags: strings.Split(r.FormValue("tags"), ","),
	}

	if err := a.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reloadAllArticles()
	http.Redirect(w, r, "/"+url, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/search/"):]
	articles, err := findArticlesByTag(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d := Data{Articles: &articles, Article: &article{}}
	t, _ := template.ParseFiles("search.html")
	t.Execute(w, &d)

}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.URL.Path = "/save/" + r.FormValue("url")
		saveHandler(w, r)
		return
	}
	t, _ := template.ParseFiles("new.html")
	t.Execute(w, &article{})
}

func reloadAllArticles() {
	articles, _ = findArticlesByTag("%")
}

func main() {

	log.Println("Start")

	var err error
	db, err = sql.Open("sqlite3", "development.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	/*
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS articles (id INTEGER PRIMARY KEY, title TEXT UNIQUE, body TEXT, tags TEXT, url TEXT UNIQUE, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
		if err != nil {
			log.Fatal(err)
		}
	*/

	reloadAllArticles()

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/new", newHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

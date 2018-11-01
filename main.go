package main

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var articles []article
var templates = template.Must(template.ParseFiles("view.html", "edit.html", "new.html", "search.html"))

type Data struct {
	Article  *article
	Articles *[]article
}

type article struct {
	Id                          int
	Title, Body, Url, TagString string
	Tags                        []string
	Created_at, Updated_at      time.Time
}

func (a *article) save() (err error) {
	if err = a.insert(); err != nil {
		err = a.update()
	}
	return
}

func (a *article) insert() (err error) {
	_, err = db.Exec("INSERT INTO articles (title, body, tags, url) VALUES (?, ?, ?, ?)", a.Title, a.Body, a.TagString, a.Url)
	return
}

func (a *article) update() (err error) {
	res, err := db.Exec("UPDATE articles SET title = ?, body = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE url = ?", a.Title, a.Body, a.TagString, a.Url)
	if r, _ := res.RowsAffected(); r == 0 {
		err = errors.New("URL cannot be empty")
	}
	return
}

func (a *article) delete() (err error) {
	_, err = db.Exec("DELETE FROM articles WHERE url = ?", a.Url)
	return
}

func (a *article) setTags() {
	a.Tags = strings.Split(a.TagString, ",")
}

func findArticleByUrl(url string) (a article, err error) {
	row := db.QueryRow("SELECT * FROM articles WHERE url = ?", url)
	err = row.Scan(&a.Id, &a.Title, &a.Body, &a.TagString, &a.Url, &a.Created_at, &a.Updated_at)
	a.setTags()
	return
}

func findArticlesByTag(tag string) ([]article, error) {
	var (
		a        article
		articles []article
	)
	m := []string{tag, "%," + tag + ",%", tag + ",%", "%," + tag}
	rows, err := db.Query("SELECT * FROM articles WHERE tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?", m[0], m[1], m[2], m[3])
	if err != nil {
		return []article{}, err
	}
	for rows.Next() {
		if err := rows.Scan(&a.Id, &a.Title, &a.Body, &a.TagString, &a.Url, &a.Created_at, &a.Updated_at); err != nil {
			return []article{}, err
		}
		a.setTags()
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return []article{}, err
	}
	return articles, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, d *Data) {
	err := templates.ExecuteTemplate(w, tmpl+".html", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]
	if url == "" {
		listHandler(w, r)
		return
	}
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	d := Data{Article: &a, Articles: &articles}
	renderTemplate(w, "view", &d)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/search/%"
	searchHandler(w, r)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/edit/"):]
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	d := Data{Article: &a}
	renderTemplate(w, "edit", &d)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/save/"):]
	a := article{
		Title:     r.FormValue("title"),
		Body:      r.FormValue("body"),
		Url:       url,
		TagString: r.FormValue("tags"),
	}
	if err := a.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.setTags()
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
	d := Data{Articles: &articles}
	renderTemplate(w, "search", &d)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.URL.Path = "/save/" + r.FormValue("url")
		saveHandler(w, r)
		return
	}
	renderTemplate(w, "new", &Data{})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/delete/"):]
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.delete()
	reloadAllArticles()
	http.Redirect(w, r, "/", http.StatusFound)
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
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS articles (id INTEGER PRIMARY KEY, title TEXT UNIQUE CHECK(title != ''), body TEXT CHECK(body != ''), tags TEXT CHECK(tags != ''), url TEXT UNIQUE CHECK(url != ''), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
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
	http.HandleFunc("/delete/", deleteHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

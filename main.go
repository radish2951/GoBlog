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

func (a *article) save() error {
	err := a.insert()
	if err != nil {
		err = a.update()
	}
	return err
}

func (a *article) insert() error {
	_, err := db.Exec("INSERT INTO articles (title, body, tags, url) VALUES (?, ?, ?, ?)", a.Title, a.Body, a.TagString, a.Url)
	return err
}

func (a *article) update() error {
	res, err := db.Exec("UPDATE articles SET title = ?, body = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE url = ?", a.Title, a.Body, a.TagString, a.Url)
	if err != nil {
		return err
	} else if r, _ := res.RowsAffected(); r == 0 {
		err = errors.New("URL cannot be empty")
	}
	return err
}

func (a *article) delete() error {
	_, err := db.Exec("DELETE FROM articles WHERE url = ?", a.Url)
	return err
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

func mustFindArticleByUrl(w http.ResponseWriter, url string) article {
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return article{}
	}
	return a
}

func mustFindArticlesByTag(w http.ResponseWriter, tag string) []article {
	articles, err := findArticlesByTag(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return []article{}
	}
	return articles
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
		r.URL.Path = "/search/%"
		searchHandler(w, r)
		return
	}
	a := mustFindArticleByUrl(w, url)
	d := Data{Article: &a, Articles: &articles}
	renderTemplate(w, "view", &d)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/edit/"):]
	a := mustFindArticleByUrl(w, url)
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
	articles := mustFindArticlesByTag(w, url)
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
	a := mustFindArticleByUrl(w, url)
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

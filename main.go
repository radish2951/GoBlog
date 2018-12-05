package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db          *sql.DB
	articles    []*article
	templates   = template.Must(template.ParseFiles("html/view.html", "html/edit.html", "html/new.html", "html/tag.html", "html/login.html", "html/header.html", "html/footer.html"))
	hash        = "$2a$10$bOcu63.qsVSgzAB0UWC3G.4qNYHyfFm4ZsuigwTq4m7Q9DSrUtUmC"
	sessionHash []byte
)

type Data struct {
	Article       *article
	Articles      []*article
	Authenticated bool
}

type article struct {
	Id                          int
	Title, Body, Url, TagString string
	Tags                        []string
	CreatedAt, UpdatedAt        time.Time
}

func (a *article) save() error {
	err := a.insert()
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			err = a.update()
		}
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

func findArticleByUrl(url string) (a *article, err error) {
	a = &article{}
	row := db.QueryRow("SELECT * FROM articles WHERE url = ?", url)
	err = row.Scan(&a.Id, &a.Title, &a.Body, &a.TagString, &a.Url, &a.CreatedAt, &a.UpdatedAt)
	a.setTags()
	a.CreatedAt = a.CreatedAt.Local()
	a.UpdatedAt = a.UpdatedAt.Local()
	return
}

func findArticlesByTag(tag string) ([]*article, error) {
	var (
		a        article
		articles []*article
	)
	m := []string{tag, "%," + tag + ",%", tag + ",%", "%," + tag}
	rows, err := db.Query("SELECT * FROM articles WHERE tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ? ORDER BY id DESC", m[0], m[1], m[2], m[3])
	if err != nil {
		return []*article{}, err
	}
	for rows.Next() {
		if err := rows.Scan(&a.Id, &a.Title, &a.Body, &a.TagString, &a.Url, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return []*article{}, err
		}
		a.setTags()
		a.CreatedAt = a.CreatedAt.Local()
		a.UpdatedAt = a.UpdatedAt.Local()
		ta := a
		articles = append(articles, &ta)
	}
	if err := rows.Err(); err != nil {
		return []*article{}, err
	}
	return articles, nil
}

func mustFindArticleByUrl(w http.ResponseWriter, url string) (*article, error) {
	a, err := findArticleByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return &article{}, err
	}
	return a, nil
}

func mustFindArticlesByTag(w http.ResponseWriter, tag string) ([]*article, error) {
	articles, err := findArticlesByTag(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return []*article{}, err
	}
	return articles, nil
}

func reloadAllArticles() {
	articles, _ = findArticlesByTag("%")
}

func renderTemplate(w http.ResponseWriter, tmpl string, d *Data) {
	err := templates.ExecuteTemplate(w, tmpl+".html", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func auth(name, password string) bool {
	res := bcrypt.CompareHashAndPassword([]byte(hash), []byte(name+password))
	return res == nil
}

func authBeforeHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authenticated(r) {
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			handler(w, r)
		}
	}
}

func generateSessionId(w http.ResponseWriter, l int) {
	b := make([]byte, l)
	rand.Read(b)
	sessionId := base64.StdEncoding.EncodeToString(b)
	sessionHash, _ = bcrypt.GenerateFromPassword([]byte(sessionId), bcrypt.DefaultCost)
	cookie := http.Cookie{
		Name:  "sessionId",
		Value: sessionId,
	}
	http.SetCookie(w, &cookie)
}

func authenticated(r *http.Request) bool {
	cookie, err := r.Cookie("sessionId")
	if err != nil || cookie.Value == "" {
		return false
	}
	sessionId := cookie.Value
	err = bcrypt.CompareHashAndPassword(sessionHash, []byte(sessionId))
	return err == nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]
	if url == "" {
		r.URL.Path = "/tag/%"
		tagHandler(w, r)
		return
	}
	a, err := mustFindArticleByUrl(w, url)
	if err != nil {
		return
	}
	d := Data{Article: a, Authenticated: authenticated(r)}
	renderTemplate(w, "view", &d)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/edit/"):]
	a, err := mustFindArticleByUrl(w, url)
	if err != nil {
		return
	}
	d := Data{Article: a}
	renderTemplate(w, "edit", &d)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/save/"):]
	a := &article{
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

func tagHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/tag/"):]
	articles, err := mustFindArticlesByTag(w, url)
	if err != nil {
		return
	}
	d := Data{Articles: articles}
	renderTemplate(w, "tag", &d)
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
	a, err := mustFindArticleByUrl(w, url)
	if err != nil {
		return
	}
	a.delete()
	reloadAllArticles()
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		password := r.FormValue("password")
		if auth(name, password) {
			generateSessionId(w, 100)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	} else if r.Method == "GET" {
		renderTemplate(w, "login", &Data{})
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:  "sessionId",
		Value: "",
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleFunc(m map[string](http.HandlerFunc)) {
	for path, handlerFunc := range m {
		http.HandleFunc(path, handlerFunc)
	}
}

func main() {

	log.Println("Start")

	var err error
	db, err = sql.Open("sqlite3", "development.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS articles (id INTEGER PRIMARY KEY, title TEXT UNIQUE CHECK(title != ''), body TEXT CHECK(body != ''), tags TEXT CHECK(tags != ''), url TEXT UNIQUE CHECK(url != ''), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatal(err)
	}

	reloadAllArticles()

	handleFunc(map[string](http.HandlerFunc){
		"/":        viewHandler,
		"/edit/":   authBeforeHandler(editHandler),
		"/save/":   authBeforeHandler(saveHandler),
		"/tag/":    tagHandler,
		"/new":     authBeforeHandler(newHandler),
		"/delete/": authBeforeHandler(deleteHandler),
		"/login":   loginHandler,
		"/logout":  logoutHandler,
	})

	http.Handle("/stylesheet/", http.StripPrefix("/stylesheet/", http.FileServer(http.Dir("stylesheet"))))
	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("image"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

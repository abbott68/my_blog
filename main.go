package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // 导入MySQL数据库驱动
)

// 博客文章结构体
type BlogPost struct {
	ID      int
	Title   string
	Content string
	User    *User
	IsLogin bool
}
type BlogPosts []BlogPost

// 用户结构体
type User struct {
	ID       int
	Username string
	Password string
	IsLogin  bool // 新增的 IsLogin 字段
}

var db *sql.DB // 全局数据库连接对象

func initDB() {
	// 连接数据库
	database, err := sql.Open("mysql", "dev:dev123456@tcp(81.70.196.2:3306)/blog_db")
	if err != nil {
		log.Fatal(err)
	}

	// 设置全局数据库连接对象
	db = database

	// 创建博客文章表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS blog_posts (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	// 创建用户表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(50) NOT NULL UNIQUE,
        password CHAR(64) NOT NULL
    )`)
	if err != nil {
		log.Fatal(err)
	}

}

// 博客文章列表
var blogPosts = []BlogPost{
	{Title: "Blog Post 1", Content: "This is the content of Blog Post 1."},
	{Title: "Blog Post 2", Content: "This is the content of Blog Post 2."},
	{Title: "Blog Post 3", Content: "This is the content of Blog Post 3."},
}

// 主页处理函数
func homeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM blog_posts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	blogPosts := []BlogPost{}
	for rows.Next() {
		post := BlogPost{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			log.Fatal(err)
		}
		blogPosts = append(blogPosts, post)
	}

	for i := range blogPosts {
		blogPosts[i].User = &User{IsLogin: true}
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		BlogPosts []BlogPost
		IsLogin   bool
	}{
		BlogPosts: blogPosts,
		IsLogin:   true, // Set the value based on the user's login status
	}

	err = tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		log.Fatal(err)
	}
}

// 创建文章页面显示处理函数
func createPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, "")
	if err != nil {
		log.Fatal(err)
	}
}

// 创建文章处理函数
func createHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := db.Exec("INSERT INTO blog_posts (title, content) VALUES (?, ?)", title, content)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// 用户注册页面显示处理函数
func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, "")
	if err != nil {
		log.Fatal(err)
	}
}

// 用户注册处理函数
func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	rawPassword := r.FormValue("password")

	hasher := sha256.New()
	hasher.Write([]byte(rawPassword))
	passwordHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// 用户登录页面显示处理函数
func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, "")
	if err != nil {
		log.Fatal(err)
	}
}

// 用户登录处理函数
func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	rawPassword := r.FormValue("password")

	var dbUser User
	err := db.QueryRow("SELECT * FROM users WHERE username=?", username).Scan(&dbUser.ID, &dbUser.Username, &dbUser.Password)
	if err != nil {
		log.Println("查询出错：", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	hasher := sha256.New()
	hasher.Write([]byte(rawPassword))
	passwordHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	if dbUser.Password == passwordHash {
		http.SetCookie(w, &http.Cookie{
			Name:  "user",
			Value: dbUser.Username,
			Path:  "/",
		})
		dbUser.IsLogin = true
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			log.Fatal(err)
		}
		err = tmpl.Execute(w, "Invalid username or password.")
		if err != nil {
			log.Fatal(err)
		}
	}
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	defer db.Close()

	//http.HandleFunc("/", homeHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createPageHandler)
	http.HandleFunc("/create/post", createHandler)
	http.HandleFunc("/register", registerPageHandler)
	http.HandleFunc("/register/post", registerHandler)
	http.HandleFunc("/login", loginPageHandler)
	http.HandleFunc("/login/auth", loginHandler)

	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

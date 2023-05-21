package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"

	_ "github.com/go-sql-driver/mysql"
)

type BlogPost struct {
	ID      int
	Title   string
	Content string
	User    *User
	IsLogin bool
}

type User struct {
	ID       int
	Username string
	Password string
	IsLogin  bool
}

type Comment struct {
	ID      int
	PostID  int
	Content string
}

var db *sql.DB

func initDB() {
	database, err := sql.Open("mysql", "dev:dev123456@tcp(81.70.196.2:3306)/blog_db")
	if err != nil {
		log.Fatal(err)
	}

	db = database

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS blog_posts (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(50) NOT NULL UNIQUE,
        password CHAR(64) NOT NULL
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (
        id INT AUTO_INCREMENT PRIMARY KEY,
        post_id INT,
        content TEXT,
        FOREIGN KEY (post_id) REFERENCES blog_posts(id)
    )`)
	if err != nil {
		log.Fatal(err)
	}
}

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
		IsLogin:   true,
	}

	err = tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		log.Fatal(err)
	}
}

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

func createHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := db.Exec("INSERT INTO blog_posts (title, content) VALUES (?, ?)", title, content)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

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

// 发布博客页面显示处理函数
func publishPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/publish.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, "")
	if err != nil {
		log.Fatal(err)
	}
}

// 发布博客处理函数
func publishHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := db.Exec("INSERT INTO blog_posts (title, content) VALUES (?, ?)", title, content)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.FormValue("postID")
	content := r.FormValue("content")
	_, err := db.Exec("INSERT INTO comments (post_id, content) VALUES (?, ?)", postID, content)
	if err != nil {
		log.Fatal(err)
	}

	// 发送邮件通知评论
	sendEmailNotification(postID, content)

	http.Redirect(w, r, "/", http.StatusFound)
}
func sendEmailNotification(postID, commentContent string) {
	// 邮件配置
	smtpHost := "smtp.example.com"
	smtpPort := "587"
	smtpUsername := "your-email@example.com"
	smtpPassword := "your-password"
	// 构建邮件内容
	to := "recipient@example.com"
	subject := "New Comment on Post #" + postID
	body := "A new comment has been posted on post #" + postID + ":\n\n" + commentContent

	// 发送邮件
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, []string{to}, []byte("Subject: "+subject+"\r\n\r\n"+body))
	if err != nil {
		log.Println("邮件发送失败:", err)
	} else {
		log.Println("邮件发送成功")
	}
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/create", createPageHandler)
	http.HandleFunc("/create/post", createHandler)
	http.HandleFunc("/register", registerPageHandler)
	http.HandleFunc("/register/post", registerHandler)
	http.HandleFunc("/login", loginPageHandler)
	http.HandleFunc("/login/auth", loginHandler)
	http.HandleFunc("/publish", publishPageHandler)
	http.HandleFunc("/publish/post", publishHandler)
	http.HandleFunc("/comment/post", commentHandler) // 添加评论处理函数

	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

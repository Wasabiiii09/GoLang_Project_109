package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"os"
	
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type Comment struct {
	ID uint `gorm:"primaryKey"`
	Text string	`gorm:"not null"`
	UserID uint `gorm:"not null"`
	User User `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
}

type PageData struct {
    User     User
    Comments []Comment
}

func page_404(w http.ResponseWriter,r *http.Request) {
	tmpl, err := template.ParseFiles("public/404.html")
	if err != nil {
		http.Error(w, "HTML not found", http.StatusSeeOther)
		return
	}
	tmpl.Execute(w, nil)
}

func clearUserCookie(w http.ResponseWriter) {
    cookie := http.Cookie{
        Name:     "user_session",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,               
        Expires:  time.Unix(0, 0),    
        HttpOnly: true,
    }
    http.SetCookie(w, &cookie)
}

func home_page(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Redirect(w, r, "/NotFound/", http.StatusSeeOther)
        return 
    }
    
    var currentUser User
    var isLoggedIn bool

    cookie, err := r.Cookie("user_session")
	if err == nil && cookie.Value != "" { // <-- Prüft, dass der Cookie nicht leer ist
    	result := db.Where("name = ?", cookie.Value).First(&currentUser)
    	if result.Error == nil {
        	isLoggedIn = true
    	}
	}

    var AllComments []Comment
    db.Preload("User").Order("created_at desc").Find(&AllComments)

    data := struct {
        User       User
        Comments   []Comment
        IsLoggedIn bool
    }{
        User:       currentUser,
        Comments:   AllComments,
        IsLoggedIn: isLoggedIn,
    }

    home_tmpl, err := template.ParseFiles("public/home.html")
    if err != nil {
        http.Error(w, "HTML nicht gefunden", http.StatusInternalServerError)
        return
    }
    
    home_tmpl.Execute(w, data)
}

func login_page(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/login/" {
        http.Redirect(w, r, "/NotFound/", http.StatusSeeOther)
        return 
    }

    w.Header().Set("Cache-Control", "no-cache, private, max-age=0, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")

    cookie, err := r.Cookie("user_session")
    if err == nil && cookie.Value != "" {
        var currentUser User
        if err := db.Where("name = ?", cookie.Value).First(&currentUser).Error; err == nil {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }
    }

    clearUserCookie(w)

    login_tmpl, err := template.ParseFiles("public/login.html")
    if err != nil {
        http.Error(w, "HTML not found", http.StatusInternalServerError)
        return 
    }
    login_tmpl.Execute(w, nil)
}

func login_function(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost{
		inputname := r.FormValue("name_input_login")
		inputpassword := r.FormValue("password_input_login")

		var FoundUser User
		result := db.Where("name = ?", inputname).First(&FoundUser)

		if result.Error != nil {
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return 
		}

		if FoundUser.Password != inputpassword {
			http.Error(w, "Falscher Password", http.StatusBadRequest)
			return
		}

		cookie := http.Cookie{
			Name: "user_session",
			Value: FoundUser.Name,
			Path: "/",
			Expires: time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func register_page(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/register/" {
        http.Redirect(w, r, "/NotFound/", http.StatusSeeOther)
        return 
    }

    w.Header().Set("Cache-Control", "no-cache, private, max-age=0, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")

    cookie, err := r.Cookie("user_session")
    if err == nil && cookie.Value != "" {
        var currentUser User
        if err := db.Where("name = ?", cookie.Value).First(&currentUser).Error; err == nil {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }
    }

    clearUserCookie(w)

    register_tmpl, err := template.ParseFiles("public/register.html")
    if err != nil {
        http.Error(w, "HTML not found", http.StatusInternalServerError)
        return
    }
    register_tmpl.Execute(w, nil)
}

func register_function(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		inputName := r.FormValue("name_input_register")
		inputPassword := r.FormValue("password_input_register")

		CreatUser := User{Name: inputName, Password: inputPassword}

		result := db.Create(&CreatUser)
		if result.Error != nil {
			http.Error(w, "Name ist besetzt", http.StatusBadRequest)
			return
		}

		cookie := http.Cookie{
			Name: "user_session",
			Value: CreatUser.Name,
			Path: "/",
			Expires: time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func log_out_function(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie := http.Cookie{
		Name: "user_session",
		Value: "",
		Path: "/",
		MaxAge: -1,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func profile_page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile/" {
		http.Redirect(w, r, "/NotFound/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("user_session")
	if err != nil {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
		return
	}

	var currentUser User 
	result := db.Where(&User{Name: cookie.Value}).First(&currentUser)

	if result.Error != nil {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
		return
	}

	profile_tmpl, err := template.ParseFiles("public/profile.html")
	if err != nil {
		http.Error(w, "HTML not found", http.StatusSeeOther)
		return
	}

	profile_tmpl.Execute(w, currentUser)

}

func comment_page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/comment/" {
		http.Redirect(w, r, "/NotFound/", http.StatusSeeOther)
		return
	}
    
    cookie, err := r.Cookie("user_session")
    if err != nil || cookie.Value == "" { // <-- Falls kein Cookie oder leerer Cookie
        http.Redirect(w, r, "/login/", http.StatusSeeOther)
        return
    }
    
    var currentUser User 
    result := db.Where("name = ?", cookie.Value).First(&currentUser)
    
	if result.Error != nil {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
		return
	}

	comment_tmpl, err := template.ParseFiles("public/comment.html")
	if err != nil {
		http.Error(w, "HTML not found", http.StatusSeeOther)
		return
	}
	
	comment_tmpl.Execute(w, currentUser)
}

func comment_function(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		inputtext := r.FormValue("comment_text_input")

		cookie, err := r.Cookie("user_session")
		if err != nil {
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return
		}

		var currentUser User
		userResult := db.Where(&User{Name: cookie.Value}).First(&currentUser)

		if userResult.Error != nil {
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return
		}	
		
		var commentCount int64
		db.Model(&Comment{}).Where("user_id = ?", currentUser.ID).Count(&commentCount)

		if commentCount >= 5 {
			http.Error(w, "You can maximum write 5 Comments", http.StatusBadRequest)
			return
		}

		CreatComment := Comment{
			Text: inputtext,
			UserID: currentUser.ID,
		}

		result := db.Create(&CreatComment)
		if result.Error != nil {
			http.Error(w, "Problem with Create Comment", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Kein .env configuiert.")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN nicht gefunden.")
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Fehler mit DB: %v", err)
	}

	db.AutoMigrate(&User{}, &Comment{})

	http.HandleFunc("/", home_page)
	http.HandleFunc("/NotFound/", page_404)
	http.HandleFunc("/login/", login_page)
	http.HandleFunc("/register/", register_page)
	http.HandleFunc("/profile/", profile_page)
	http.HandleFunc("/comment/", comment_page)

	http.HandleFunc("/register_action", register_function)
	http.HandleFunc("/login_action", login_function)

	http.HandleFunc("/comment_action", comment_function)

	http.HandleFunc("/log_out", log_out_function)

	fmt.Println("Server: http://localhost:7777")
	log.Fatal(http.ListenAndServe(":7777", nil))
}

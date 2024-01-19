package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Data struct {
	UserID       string `json:"userid,omitempty"`
	FirstName    string `json:"firstname,omitempty"`
	LastName     string `json:"lastname,omitempty"`
	Email        string `json:"email,omitempty"`
	Phonenumber  string `json:"phonenumber,omitempty"`
	Postalcode   string `json:"postalcode,omitempty"`
	Housenumber  string `json:"housenumber,omitempty"`
	Street       string `json:"street,omitempty"`
	Town         string `json:"town,omitempty"`
	Country      string `json:"country,omitempty"`
	Birthdate    string `json:"birthdate,omitempty"`
	Licenseplate string `json:"licenseplate,omitempty"`
}

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Database string `json:"database"`
}

func getconfig() (string, string, string, string, string) {
	// Read the content of the aconfig.json file
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		fmt.Println("Error reading config.json:", err)
		return "", "", "", "", ""
	}

	// Parse the JSON data into the Config struct
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", "", "", "", ""
	}
	return config.Username, config.Password, config.Ip, config.Port, config.Database
}

func main() {
	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")
	username, password, ip, port, database := getconfig()
	connectionstring := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=true", username, password, ip, port, database)
	db, err := sql.Open("mysql", connectionstring)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/licenseplate", func(w http.ResponseWriter, r *http.Request) {
		licenseplate := r.URL.Query().Get("licenseplate")
		if licenseplate != "" {
			info := db.QueryRow("SELECT firstname FROM user INNER JOIN reservering ON user.userid = reservering.userid WHERE licenseplate = ? AND checkout >= ? AND checkin <= ?", licenseplate, currentDate, currentDate)
			var data Data
			err = info.Scan(&data.FirstName)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "licenseplate not allowed", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		} else {
			http.Error(w, "Please enter an licenseplate", http.StatusNotFound)
		}
	})

	http.HandleFunc("/reservering", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		case http.MethodPost:
			checkin := r.URL.Query().Get("checkin")
			checkout := r.URL.Query().Get("checkout")
			housenumber := r.URL.Query().Get("housenumber")
			email := r.URL.Query().Get("email")
			password := r.URL.Query().Get("password")

			if checkin != "" && checkout != "" && email != "" && housenumber != "" && password != "" {
				var hashedPassword string
				err = db.QueryRow("SELECT password FROM user WHERE email=?", email).Scan(&hashedPassword)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "email or password not valid", http.StatusNotFound)
						return
					}
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
				if err != nil {
					http.Error(w, "email or password not valid", http.StatusNotFound)
					return
				}
				db.QueryRow("INSERT INTO reservering (userid, checkin, checkout, housenumber) VALUES ((SELECT userid FROM user WHERE email = ?), ?, ?, ?)", email, checkin, checkout, housenumber)
				if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
				http.Error(w, "Reservation added", http.StatusOK)
			} else {
				http.Error(w, "Please enter checkin, checkout, email, password and housenumber", http.StatusNotFound)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	})

	http.HandleFunc("/user/add", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		case http.MethodPost:
			firstname := r.URL.Query().Get("firstname")
			lastname := r.URL.Query().Get("lastname")
			email := r.URL.Query().Get("email")
			password := r.URL.Query().Get("password")
			phonenumber := r.URL.Query().Get("phonenumber")
			postalcode := r.URL.Query().Get("postalcode")
			housenumber := r.URL.Query().Get("housenumber")
			street := r.URL.Query().Get("street")
			town := r.URL.Query().Get("town")
			country := r.URL.Query().Get("country")
			birthdate := r.URL.Query().Get("birthdate")
			licenseplate := r.URL.Query().Get("licenseplate")

			if firstname != "" && lastname != "" && email != "" && password != "" && postalcode != "" && housenumber != "" && street != "" && town != "" && country != "" && birthdate != "" && licenseplate != "" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "Failed to hash password", http.StatusInternalServerError)
					return
				}
				info := db.QueryRow("SELECT email FROM user WHERE email=?", email)
				var data Data
				err = info.Scan(&data.Email)
				if err != nil {
					if err != sql.ErrNoRows {
						http.Error(w, "user with this email already excists", http.StatusConflict)
						fmt.Println(err)
						return
					}
				}
				info = db.QueryRow("SELECT licenseplate FROM user WHERE licenseplate=?", licenseplate)
				err = info.Scan(&data.Licenseplate)
				if err != nil {
					if err != sql.ErrNoRows {
						http.Error(w, "user with this licenseplate already excists", http.StatusConflict)
						fmt.Println(err)
						return
					}
				}
				db.QueryRow("INSERT INTO `reserveringen`.`user` (`firstname`, `lastname`, `email`, `password`, `phonenumber`, `postalcode`, `housenumber`, `street`, `town`, `country`, `birthdate`, `licenseplate`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", firstname, lastname, email, hashedPassword, phonenumber, postalcode, housenumber, street, town, country, birthdate, licenseplate)
				if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
				http.Error(w, "User added", http.StatusOK)
			} else {
				http.Error(w, "Please enter all data necessary", http.StatusNotFound)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	})

	http.HandleFunc("/user/modify", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		case http.MethodPost:
			firstname := r.URL.Query().Get("firstname")
			lastname := r.URL.Query().Get("lastname")
			birthdate := r.URL.Query().Get("birthdate")
			town := r.URL.Query().Get("town")

			email := r.URL.Query().Get("email")
			oldpassword := r.URL.Query().Get("oldpassword")
			newpassword := r.URL.Query().Get("newpassword")
			phonenumber := r.URL.Query().Get("phonenumber")
			licenseplate := r.URL.Query().Get("licenseplate")

			if firstname != "" && lastname != "" && birthdate != "" && town != "" {
				oldhashedPassword, err := bcrypt.GenerateFromPassword([]byte(oldpassword), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "Failed to hash password", http.StatusInternalServerError)
					return
				}
				newhashedPassword, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "Failed to hash password", http.StatusInternalServerError)
					return
				}

				info := db.QueryRow("SELECT userid, email FROM user WHERE firstname=? AND lastname=? AND birthdate=? AND town =?", firstname, lastname, birthdate, town)
				var data Data
				err = info.Scan(&data.UserID, &data.Email)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "There is no user found", http.StatusConflict)
						return
					}
				}
				if email != "" {
					db.QueryRow("UPDATE user SET email=? WHERE userid=?", email, data.UserID)
				} else if newpassword != "" {
					db.QueryRow("UPDATE user SET password=? WHERE userid=? AND password=?", newhashedPassword, data.UserID, oldhashedPassword)
				} else if phonenumber != "" {
					db.QueryRow("UPDATE user SET phonenumber=? WHERE userid=?", phonenumber, data.UserID)
				} else if licenseplate != "" {
					db.QueryRow("UPDATE user SET licenseplate=? WHERE userid=?", licenseplate, data.UserID)
				} else {
					http.Error(w, "Please enter all data necessary", http.StatusNotFound)
				}
				http.Error(w, "User modified", http.StatusOK)
			} else {
				http.Error(w, "Please enter all data necessary", http.StatusNotFound)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/user/delete", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		case http.MethodPost:
			email := r.URL.Query().Get("email")
			password := r.URL.Query().Get("password")
			if email != "" && password != "" {
				var hashedPassword string
				err = db.QueryRow("SELECT password FROM user WHERE email=?", email).Scan(&hashedPassword)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "email or password not valid", http.StatusNotFound)
						return
					}
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
				if err != nil {
					http.Error(w, "email or password not valid", http.StatusNotFound)
					return
				}
				db.QueryRow("DELETE FROM user WHERE email=?", email)
				http.Error(w, "User deleted", http.StatusOK)
			} else {
				http.Error(w, "Please enter email and password", http.StatusNotFound)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/user/get", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		password := r.URL.Query().Get("password")
		if email != "" && password != "" {
			var hashedPassword string
			err = db.QueryRow("SELECT password FROM user WHERE email=?", email).Scan(&hashedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "email or password not valid", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
			if err != nil {
				http.Error(w, "email or password not valid", http.StatusNotFound)
				return
			}
			info := db.QueryRow("SELECT firstname, lastname, email, phonenumber, postalcode, housenumber, street, town, country, birthdate, licenseplate FROM user WHERE email=?", email)
			var data Data
			err = info.Scan(&data.FirstName, &data.LastName, &data.Email, &data.Phonenumber, &data.Postalcode, &data.Housenumber, &data.Street, &data.Town, &data.Country, &data.Birthdate, &data.Licenseplate)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "email or password not valid", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		} else {
			http.Error(w, "Please enter email and password", http.StatusNotFound)
		}
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		password := r.URL.Query().Get("password")
		if email != "" && password != "" {
			var hashedPassword string
			err = db.QueryRow("SELECT password FROM user WHERE email=?", email).Scan(&hashedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "email or password not valid", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
			if err != nil {
				http.Error(w, "email or password not valid", http.StatusNotFound)
				return
			}
			http.Error(w, "Login succesful", http.StatusOK)
		} else {
			http.Error(w, "Please enter email and password", http.StatusNotFound)
		}
	})
	log.Fatal(http.ListenAndServe(":80", nil))
}

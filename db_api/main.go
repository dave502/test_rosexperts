package main

import (
"fmt"
"database/sql"
"encoding/json"
"net/http"
"log"
"os"
"io"
"strings"
"time"
socket "github.com/lxzan/gws"
_ "github.com/lib/pq"
)

var socket_url string
var db *sql.DB
var socket_conn *socket.Conn

type Statement struct {
	Func string
	Args string
}

// This function will make a connection to the database only once.
func main() {
	var err error

	dbuser     := os.Getenv("DB_USER")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname     := os.Getenv("DB_NAME")
	dbhost     := os.Getenv("DB_HOST")
	dbport     := os.Getenv("DB_PORT")

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbhost, dbport, dbuser, dbpassword, dbname)

	// open database
	for db, err = sql.Open("postgres", psqlconn); err != nil; db, err = sql.Open("postgres", psqlconn){
		log.Println(err, "Failed to open database ", psqlconn, " Keep trying...")
		time.Sleep(5 * time.Second)
	}

	for err = db.Ping(); err != nil; err = db.Ping() {
		log.Println(err, "Failed to ping database. Keep trying...")
		time.Sleep(5 * time.Second)
	}

	// this will be printed in the terminal, confirming the connection to the database
	log.Println("The database is connected")

	init_db()

	go runHTTPServer()

	socket.NewServer(new(SocketHandler),  &socket.ServerOption{
		ReadMaxPayloadSize:      64 * 1024 * 1024,
		ReadBufferSize:     64 * 1024 * 1024,
		WriteMaxPayloadSize:  64 * 1024 * 1024}).Run(":8000")
}

func runHTTPServer(){
	fmt.Printf("strarting http server on port 3333")
	http.HandleFunc("/gettext", getText)
	err := http.ListenAndServe(":3333", nil)
	if err == http.ErrServerClosed {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getText(w http.ResponseWriter, r *http.Request){
	corpus_arr := []string{}
	fmt.Printf("got / request getText\n")
	sqlStatement := `SELECT data_text FROM data`
	rows, err := db.Query(sqlStatement)
	if err != nil{
		fmt.Printf("Failed fetch data_text: %s\n", err)
	}
	for rows.Next() {
        var doc string
        rows.Scan(&doc)
        corpus_arr = append(corpus_arr, doc)
    }
	corpus := strings.Join(corpus_arr,"\n")
	io.WriteString(w, corpus)
}


type SocketHandler struct {
	socket.BuiltinEventHandler
}

func (c *SocketHandler) OnPing(s *socket.Conn, payload []byte) {
	log.Println("ping recieved")
	_ = s.WritePong(payload)
}


func (c *SocketHandler) OnMessage(s *socket.Conn, message *socket.Message) {
	defer message.Close()

	var response string
	var msg Statement

	if err := json.Unmarshal(message.Bytes(), &msg); err != nil{
		log.Println("msg", msg)
		log.Println("Failed unmarshal message", err)
	} else {
		switch msg.Func{
			case "AppendText":
				err := AppendText(msg.Args)
				if err != nil {
					log.Println(err, "Failed Append")
					response = "500"
				} else {
					response = "200"
				}
			default:
				response = "400"
		}
		// log.Println("message recieved:", string(message.Bytes()))
		// log.Println("struct:", msg)
	}

	_ = s.WriteMessage(message.Opcode, []byte(response))
}


func (c *SocketHandler) OnOpen(s *socket.Conn) {

	log.Println("connected")
}


func AppendText(text string) error {
	sqlStatement := `INSERT INTO data (id, data_text) VALUES (1, $1)
		ON CONFLICT (id) DO
		UPDATE SET data_text = data.data_text || $1;`
	_, err := db.Exec(sqlStatement, text)
	return err
}


func init_db(){
	db.Exec(`CREATE TABLE IF NOT EXISTS data (
		id serial PRIMARY KEY,
		data_updated timestamp DEFAULT current_timestamp,
		data_text text NOT NULL,
		data_vector int[][]
	)`)
}

func insert(table string, fields []string, values[]string){
	insertDynStmt := `insert into ` + table + ` ("` + strings.Join(fields,`","`) + `" values("` + strings.Join(values,`","`) +`")`
	log.Println(insertDynStmt)
    _, err := db.Exec(insertDynStmt)
	failOnError(err, "Failed Insert")
}


func failOnError(err error, msg string) {
	//log.Println(msg)
	if err != nil {
	  log.Panicf("%s: %s", msg, err)
	}
}

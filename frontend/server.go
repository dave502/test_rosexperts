package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
)

func main() {
	fmt.Printf("strarting http server on port 8888")
	http.HandleFunc("/", sendIndex)
	err := http.ListenAndServe(":8888", nil)
	if err == http.ErrServerClosed {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func sendIndex(w http.ResponseWriter, r *http.Request){

	flowise_host := os.Getenv("FLOWISE_HOST")

	indexhtml := fmt.Sprintf(`<!DOCTYPE html>
		<html lang="en">
		  <head>
		    <meta charset="UTF-8">
		    <meta name="viewport" content="width=device-width, initial-scale=1.0">
		    <meta http-equiv="X-UA-Compatible" content="ie=edge">
		    <title>Knowledge base</title>
		  </head>
		  <body>
		    <main>
			<h2>Knowledge base Rosexperts</h2>  
		    </main>
		  <script type="module">
		      import Chatbot from "https://cdn.jsdelivr.net/npm/flowise-embed/dist/web.js"
		      Chatbot.init({
			  chatflowid: "09235cba-71c9-4d92-9228-41c561b297ef",
			  apiHost: "%s",
		      })
		  </script>
		  </body>
		</html>
		`, flowise_host)
	io.WriteString(w, indexhtml)
}

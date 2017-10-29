package main

import "net"
import "fmt"
import "bufio"
import "os"
import "strings"
//import "log"

var msg string

func main() {

    // connect to this socket
    conn, err := net.Dial("tcp", "127.0.0.1:8484")
    if err != nil {
        fmt.Println("dial error:", err)
        return
    }
    fmt.Println("connected")
//    msg, _ := bufio.NewReader(conn).ReadString('\n')
//    fmt.Println("msg received")
//    fmt.Println(msg)
    // fmt.Println(msg)
    // fmt.Println("smth")
		stdin := bufio.NewReader(os.Stdin)
		tcp   := bufio.NewReader(conn)


    for { 
        // read in input from stdin
        fmt.Print("reading from stdin: ")
        text, _ := stdin.ReadString('\n')

		if (len(strings.TrimSpace(text)) >= 1){
			fmt.Printf(" sending to server: %s", text)
	        // send to socket
	        fmt.Fprintf(conn, text)
		 	fmt.Print(" server's response: ")
			msg, _ := tcp.ReadString('\n')
			trim := strings.TrimSpace(msg)
			fmt.Println(trim)

		}
		

        // listen for reply
        //message, _ := bufio.NewReader(conn).ReadString('\n')
        //fmt.Print("Message from server: "+message)
    }
}

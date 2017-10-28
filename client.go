package main

import "net"
import "fmt"
import "bufio"
import "os"

func main() {

    // connect to this socket
    conn, err := net.Dial("tcp", "127.0.0.1:8080")
    if err != nil {
        fmt.Println("dial error:", err)
        return
    }
    fmt.Println("connected")
    msg, _ := bufio.NewReader(conn).ReadString('\n')
    fmt.Println("msg received")
    fmt.Println(msg)
    // fmt.Println(msg)
    // fmt.Println("smth")
    for { 
        // read in input from stdin
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Text to send: ")
        text, _ := reader.ReadString('\n')
        fmt.Printf("input is: %s", text)
        // send to socket
        fmt.Fprintf(conn, text)
        // listen for reply
        //message, _ := bufio.NewReader(conn).ReadString('\n')
        //fmt.Print("Message from server: "+message)
    }
}

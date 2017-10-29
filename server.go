package main

import "net"
import "fmt"
import "bufio"
import "strings" // only needed below for sample processing





func main() {

  fmt.Println("Launching server...")

  for {
    ln, _ := net.Listen("tcp", "10.112.3.37:8081")
    conn, _ := ln.Accept()
                                                  // will listen for message to process ending in newline (\n)
    message, _ := bufio.NewReader(conn).ReadString('\n')
                                                     // output message received
    fmt.Print("Message Received:", string(message))
                                                  // sample process for string received
    newmessage := strings.ToUpper(message)
                                              // send new string back to client
    conn.Write([]byte(newmessage + "\n"))
  }
}

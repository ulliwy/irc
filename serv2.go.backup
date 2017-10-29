package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
)

var allClients map[*Client]int

type Client struct {
    // incoming chan string
    nickname    string
    outgoing   chan string
    reader     *bufio.Reader
    writer     *bufio.Writer
    conn       net.Conn
    connection *Client
}

func (client *Client) Read() {
    for {
        line, err := client.reader.ReadString('\n')
        if err == nil {
            if client.connection != nil {
                client.connection.outgoing <- line
            }
            fmt.Print(client.nickname, ": ", line)
        } else {
            break
        }

    }

    client.conn.Close()
    delete(allClients, client)
    if client.connection != nil {
        client.connection.connection = nil
    }
    client = nil
}

func (client *Client) Write() {
    for data := range client.outgoing {
        client.writer.WriteString(data)
        client.writer.Flush()
    }
}

func (client *Client) Listen() {
    go client.Read()
    go client.Write()
}

func NewClient(connection net.Conn) *Client {
    writer := bufio.NewWriter(connection)
    reader := bufio.NewReader(connection)


    //connection.Write([]byte("hello \n"))
    writer.WriteString("Enter your nickname: \n")
    writer.Flush()
    fmt.Println("msg sent")
    msg, _ := reader.ReadString('\n')
    client := &Client{
        // incoming: make(chan string),
        nickname:   strings.TrimSpace(msg),
        outgoing: make(chan string),
        conn:     connection,
        reader:   reader,
        writer:   writer,
    }
    fmt.Println("My name is " + client.nickname)
    client.Listen()

    return client
}

func main() {
    allClients = make(map[*Client]int)
    listener, _ := net.Listen("tcp", "127.0.0.1:8080")
    fmt.Println("Server is running...")
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println(err.Error())
        }
        client := NewClient(conn)
        for clientList, _ := range allClients {
            if clientList.connection == nil {
                client.connection = clientList
                clientList.connection = client
                fmt.Println("Connected")
            }
        }
        allClients[client] = 1
        fmt.Println(len(allClients))
    }
}
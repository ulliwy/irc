package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
)

var hostname string = "127.0.0.1"
var port string = "8484"

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
                client.connection.outgoing <- line //trimmed
            	}
			trim := strings.TrimSpace(line)
			if ( trim == ""){
				fmt.Fprintf(client.conn, ":" + hostname + " 421 <> :Unknown command\n")
				return}
			arg := strings.Fields(line)

			if (len(arg) >= 2){ //if a two word command was sent to server
				if (arg[0] == "NICK"){ //if first word is NICK
					if (client.nickname == "*"){
						client.nickname = arg[1]
						fmt.Fprintf(client.conn, ":" + hostname + " 001 " + client.nickname + " :Welcome to ft-irc-go " + client.nickname + "\n")
					} else {
						client.nickname = arg[1]
						fmt.Fprintf(client.conn, ":" + hostname + " NOTICE " + client.nickname + " :Your nickname is now " + client.nickname + "\n")
					}
				}
			}

			fmt.Println(client.nickname + " said: " + trim)
			if (trim == "CAP LS 302"){
				fmt.Fprintf(client.conn, ":" + hostname + " CAP " + client.nickname + " LS :account-notify extended-join identify-msg multi-prefix sasl\n")
			} else if (trim == "CAP REQ identify-msg"){
				fmt.Fprintf(client.conn, ":" + hostname + " CAP " + client.nickname + " ACK :identify-msg\n")
			} else if (trim == "CAP REQ multi-prefix"){
				fmt.Fprintf(client.conn, ":" + hostname + " CAP " + client.nickname + " ACK :multi-prefix\n")
			} else if (trim == "CAP REQ sasl"){
				fmt.Fprintf(client.conn, ":" + hostname + " CAP " + client.nickname + " ACK :sasl\n")
			} else if (trim == "CAP REQ userhost-in-names"){
				fmt.Fprintf(client.conn, ":" + hostname + " CAP " + client.nickname + " ACK :userhost-in-names\n")
			} else {

				fmt.Fprintf(client.conn, ":" + hostname + " 421 <" + arg[0] + "> :Unknown command\n")
			}

 
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

    client := &Client{
		nickname: "*",
        outgoing: make(chan string),
        conn:     connection,
        reader:   reader,
        writer:   writer,
    }
    client.Listen()
    return client
}

func main() {
    allClients = make(map[*Client]int)
    listener, _ := net.Listen("tcp", hostname + ":" + port )
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
            }
        }
    }
}

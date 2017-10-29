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
	username	string
	hostname	string
	servername	string
	realname	string
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
			words := strings.Fields(line)
			if (words[0] == "NICK"){ //if first word is NICK
				if words[1] == ""{ //ERR_NONICKNAMEGIVEN             = "431"
					fmt.Fprintf(client.conn, ":" + hostname + " 431 " + client.nickname + " :No Nick Name Given\n")
					break
				}
				for clientList, _ := range allClients { //// IMPORTANT - BUGGY - NEEDS TO BE FIXED
					if clientList.nickname == words[1]{   //ERR_NICKNAMEINUSE               = "433"
					fmt.Println("we have two duplicate nicknames!!!")
					fmt.Fprintf(client.conn, ":" + hostname + " 433 " + client.nickname + " :Nickname is already registered\n")
					break
					}
				}
				if (client.nickname == "*"){
						client.nickname = words[1]
					fmt.Fprintf(client.conn, ":" + hostname + " 001 " + client.nickname + " :Welcome to ft-irc-go " + client.nickname + "\n")
				} else {
					client.nickname = words[1]
					fmt.Fprintf(client.conn, ":" + hostname + " NOTICE " + client.nickname + " :Your nickname is now " + client.nickname + "\n")
				}
			}
			if (words[0] == "PING") {
				if (words[1] != ""){
					fmt.Fprintf(client.conn, ":" + hostname + " PONG " + hostname + ":" + words[1] + "\n") 
				}
			}
			if (words[0] == "USER") {
				if ( len(words) < 5){ //ERR_NEEDMOREPARAMS
					fmt.Fprintf(client.conn, ":" + hostname + " 461 " + client.nickname + " :Not enough parameters\n")
				} else if (client.username != ""){ //ERR_ALREADYREGISTRED
					fmt.Fprintf(client.conn, ":" + hostname + " 462 " + client.nickname + " :You may not reregister\n")
					} else {
					client.username = words[1]
					client.hostname = words[2]
					client.servername = words[3]
					client.realname = words[4][1:]
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
				fmt.Fprintf(client.conn, ":" + hostname + " 421 <" + words[0] + "> :Unknown command\n")
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
		username: "",
		hostname: "",
		servername: "",
		realname: "",
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

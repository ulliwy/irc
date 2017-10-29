package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
)

type ConnectedClient struct {
	conn net.Conn
	writer *bufio.Writer
	reader *bufio.Reader

	username string
	nickname string
	realname string
	registered bool

	msg_cnt int
}

// key is a nickname
var clients map[string]*ConnectedClient

func (client *ConnectedClient) respond(msg string) {
	fmt.Fprintf(client.conn, ":localhost %03d %s :%s\n", client.msg_cnt, client.nickname, msg)
	client.msg_cnt = client.msg_cnt + 1
}

func (client *ConnectedClient) error(msg NumericMessage, args ...interface{}) {
	err_str := fmt.Sprintf(msg.template, args)
	fmt.Fprintf(client.conn, ":localhost %03d %s :%s\n", msg.number, client.nickname, err_str)
}

func (client *ConnectedClient) register() {
	if client.registered {
		return
	}

	if client.username != "" && client.nickname != "" {
		client.registered = true
		client.respond("Hello")
	}
}

func (client *ConnectedClient) unregister() {
	if client.nickname != "" {
		delete(clients, client.nickname)
	}
}

func (client *ConnectedClient) user_command(username string, realname string) {
	client.username = username
	client.realname = realname
	client.register()
}

func (client *ConnectedClient) nick_command(nickname string) {
	old_nickname := client.nickname
	client.nickname = nickname

	_, ok := clients[nickname]
	if !ok {
		clients[nickname] = client

		if old_nickname != "" {
			delete(clients, old_nickname)
		}
	} else {
		client.error(ERR_NICKNAMEINUSE)
		client.nickname = old_nickname
		return
	}

	client.register()
}

func (client *ConnectedClient) privmsg_command(nickname string, text string) {
	if !client.registered {
		client.error(ERR_NOTREGISTERED)
		return
	}

	receiver_client, ok := clients[nickname]
	if ok {
		receiver_client.respond(text)
	} else {
		client.error(ERR_NOSUCHNICK, nickname)
	}
}

func (client *ConnectedClient) handle_command(cmd string) {
	args := strings.Split(cmd, " ")

	fisrt_cmd := args[0]
	if fisrt_cmd == "USER" {
		client.user_command(args[1], args[4])
	} else if fisrt_cmd == "NICK" {
		client.nick_command(args[1])
	} else if fisrt_cmd == "PRIVMSG" {
		//var receivers []string
		//receivers = append(receivers, args[1])
		client.privmsg_command(args[1], args[2])
	} else {
		if (client.registered) {
			client.error(ERR_UNKNOWNCOMMAND)
		}
	}

	fmt.Println("$")
	for key, _ := range clients {
    	fmt.Printf("Key: %s\n", key)
	}
	fmt.Println("$")
}

func handle_conn(conn net.Conn) {
	client := ConnectedClient {
		conn: conn,
		msg_cnt: 1,
	}

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		fmt.Print(msg)
		client.handle_command(strings.TrimSuffix(msg, "\n"))
		client.handle_command(strings.TrimSuffix(msg, "\r\n"))
	}

	client.unregister()
	conn.Close()
}

func main() {
	clients = make(map[string]*ConnectedClient)

	fmt.Println("Hello, 世界")
	listener, _ := net.Listen("tcp", "127.0.0.1:6668")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error")
			continue
		}
		go handle_conn(conn)
	}
}
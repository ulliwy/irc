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
	hostname string
	servername string
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

func (client *ConnectedClient) user_command(username string, hostname string, servername string, realname string) {
	client.username = username
	client.realname = realname
	client.hostname = hostname
	client.servername = servername
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

	fisrt_cmd := ""
	len := len(args)
	if len > 0 {
		fisrt_cmd = strings.ToUpper(args[0])
	}
	if fisrt_cmd == "USER" {
		if len >= 5 {
			rname := ""
			for i := 4; i < len; i++ {
				rname += args[i]
				rname += " "
			}
			rname = strings.TrimSuffix(rname, " ")
			rname = strings.TrimPrefix(rname, ":")
			client.user_command(args[1], args[2], args[3], rname)
		}
	} else if fisrt_cmd == "NICK" {
		if len >= 1 {
			client.nick_command(args[1])
		}
	} else if fisrt_cmd == "PRIVMSG" {
		if len >= 3 {
			msg := ""
			for i := 2; i < len; i++ {
				msg += args[i]
				msg += " "
			}
			fmt.Printf("msg: {%s}\n", msg)
			client.privmsg_command(args[1], strings.TrimSuffix(msg, " "))
		}
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
		msg = strings.TrimSuffix(msg, "\r\n")
		msg = strings.TrimSuffix(msg, "\n")
		client.handle_command(msg)
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
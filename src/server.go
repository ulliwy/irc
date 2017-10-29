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
	channels map[string]string

	msg_cnt int
}

type Channel struct {
	name string
	chan_clients map[string]*ConnectedClient
}

var clients map[string]*ConnectedClient
var channels map[string]*Channel


func (client *ConnectedClient) respond_channel(msg string, channel_name string, person *ConnectedClient) {
	fmt.Fprintf(person.conn, ":%s!~%s@%s PRIVMSG %s %s\n", client.nickname, client.username, client.hostname, channel_name, msg)
}

func (client *ConnectedClient) respond_peer(msg string, peer *ConnectedClient) {
	fmt.Fprintf(peer.conn, ":%s!~%s@%s PRIVMSG %s %s\n", client.nickname, client.username, client.hostname, peer.username, msg)
	client.msg_cnt = client.msg_cnt + 1
}

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
		client.respond_peer(text, receiver_client)
	} else {
		fmt.Printf("here\n")
		client.error(ERR_NOSUCHNICK, nickname)
	}
}

func (client *ConnectedClient) msg_to_channel(channel_name string, text string) {
	if !client.registered {
		client.error(ERR_NOTREGISTERED)
		return
	}

	receiver_channel, ok := channels[channel_name]
	if ok {
		for key:= range receiver_channel.chan_clients {
			person := receiver_channel.chan_clients[key]
			if person.username != client.username {
				fmt.Printf("client:{%s}, person{%s}", client.nickname, person.nickname)
				client.respond_channel(text, channel_name, person)
			}
		}
	} else {
		fmt.Printf("here1\n")
		client.error(ERR_NOSUCHNICK, channel_name)
	}
}

func (client *ConnectedClient) join_channel(channel_name string) {
	if !client.registered {
		client.error(ERR_NOTREGISTERED)
		return
	}
	fmt.Printf("{%s}\n", channel_name)
	ch, ok := channels[channel_name]
	if ok {
		_, ok := ch.chan_clients[client.username]
		if !ok {
			ch.chan_clients[client.username] = client
		}
	} else {
		channel := Channel {
			name: channel_name,
			chan_clients: make(map[string]*ConnectedClient),
		}
		channels[channel_name] = &channel
		channel.chan_clients[client.username] = client
		client.channels[channel_name] = channel_name
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
			if args[1][0] == '#' {
				client.msg_to_channel(args[1], strings.TrimSuffix(msg, " "))
			} else {
				client.privmsg_command(args[1], strings.TrimSuffix(msg, " "))
			}
		}
	} else if fisrt_cmd == "JOIN" {
		if len > 1 {
			chan_lst := strings.Split(args[1], ",")
			for i := range chan_lst {
				client.join_channel(chan_lst[i])
			}
		}
	} else {
		if (client.registered) {
			client.error(ERR_UNKNOWNCOMMAND)
		}
	}

	fmt.Println("$")
	for key, _ := range channels {
    	fmt.Printf("Key: %s\n", key)
	}
	fmt.Println("$")
}

func handle_conn(conn net.Conn) {
	client := ConnectedClient {
		conn: conn,
		msg_cnt: 1,
		channels: make(map[string]string),
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
	channels = make(map[string]*Channel)

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
package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"time"
)

var hostname string = "127.0.0.1"
var port     string = "6667"

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
	created int32
//	subject string
}

var clients map[string]*ConnectedClient
var channels map[string]*Channel


func (client *ConnectedClient) respond_channel(msg string, channel_name string, person *ConnectedClient) {
	fmt.Fprintf(person.conn, ":%s!~%s@%s PRIVMSG %s %s\r\n", client.nickname, client.username, client.hostname, channel_name, msg)
}

func (client *ConnectedClient) respond_peer(msg string, peer *ConnectedClient) {
	fmt.Fprintf(peer.conn, ":%s!~%s@%s PRIVMSG %s %s\r\n", client.nickname, client.username, client.hostname, peer.username, msg)
	client.msg_cnt = client.msg_cnt + 1
}

func (client *ConnectedClient) respond(msg string) {
	fmt.Fprintf(client.conn, ":" + hostname + " %03d %s :%s\r\n", client.msg_cnt, client.nickname, msg)
	client.msg_cnt = client.msg_cnt + 1
}

func (client *ConnectedClient) respond_raw(msg string) {
		fmt.Fprintf(client.conn, msg)
}

func (client *ConnectedClient) cap_command(args []string){

	if (args[1] == "LS"){
	client.respond_raw(":" + hostname + " CAP * LS : identify-msg\r\n")
	} else if (args[1] == "REQ") && (args[2] != ""){
			if (args[2] == "identify-msg"){
				client.respond_raw(":" + hostname + " CAP " + client.nickname + " ACK :identify-msg\r\n")
			} else if (args[2] == "END") {

			} else if (args[2] != ""){
				client.respond_raw(":" + hostname + " CAP " + client.nickname + " NAK :" + args[2] + "\r\n")
			}
		}
}


func (client *ConnectedClient) error(msg NumericMessage, args ...interface{}) {
	err_str := fmt.Sprintf(msg.template, args)
	fmt.Fprintf(client.conn, ":" + hostname + " %03d %s :%s\r\n", msg.number, client.nickname, err_str)
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
			created: int32(time.Now().Unix()),
		}
		channels[channel_name] = &channel
		channel.chan_clients[client.username] = client
		client.channels[channel_name] = channel_name
	}


	receiver_channel, ok := channels[channel_name]
	if ok {
		for key := range receiver_channel.chan_clients {
			person := receiver_channel.chan_clients[key]
			fmt.Fprintf(person.conn, ":%s!~%s@%s JOIN %s \r\n", client.nickname, client.username, client.hostname, channel_name)
			}
		}
}

func (client *ConnectedClient) who_command(args []string){
	receiver_channel, ok := channels[args[1]]
	if ok {
		for key := range receiver_channel.chan_clients {
			person := receiver_channel.chan_clients[key]
			person_info := person.username + " " + person.hostname + " " + hostname + " " +  person.nickname + " H :0 " + person.realname
					
			fmt.Fprintf(client.conn, ":" + hostname + " 352 " + client.nickname + " " + args[1] + " " + person_info + "\r\n");
		}
			fmt.Fprintf(client.conn, ":" + hostname + " 315 " + client.nickname + " " + args[1] + " :End of /WHO list.\r\n");
	}


}

func (client *ConnectedClient) names_command(args []string){
	receiver_channel, ok := channels[args[1]]
	var names string
	names = ""
	if ok {
		for key := range receiver_channel.chan_clients {
			person := receiver_channel.chan_clients[key]
			names = names + person.nickname + " "
			}
			fmt.Fprintf(client.conn, ":" + hostname + " 353 " + client.nickname + " = " + args[1] + " :" + names + "\r\n");
			fmt.Fprintf(client.conn, ":" + hostname + " 366 " + client.nickname + " " + args[1] + " :End of /NAMES list.\r\n");
		}
}

func (client *ConnectedClient) names2_command(chn string){
		var names string
		for key := range channels[chn].chan_clients {
			person := channels[chn].chan_clients[key]
			names = names + person.nickname + " "
			}
			fmt.Fprintf(client.conn, ":" + hostname + " 353 " + client.nickname + " = " + chn + " :" + names + "\r\n");
			fmt.Fprintf(client.conn, ":" + hostname + " 366 " + client.nickname + " " + chn + " :End of /NAMES list.\r\n");
}


func (client *ConnectedClient) part_command(args []string){

	quit_msg := strings.TrimPrefix(strings.Join(args[2:], " "), ":")

	receiver_channel, ok := channels[args[1]]
	if ok {
		for key := range receiver_channel.chan_clients {
			person := receiver_channel.chan_clients[key]
			fmt.Fprintf(person.conn,":%s!~%s@%s PART %s :\"%s\"\r\n",client.nickname,client.username,client.hostname,args[1],quit_msg)
				}
			}
		delete(receiver_channel.chan_clients, client.username)
		delete(client.channels, receiver_channel.name)

		for key, _ := range channels {
			if (len(channels[key].chan_clients) == 0){
				delete(channels, key)
			}
		}
}


func (client *ConnectedClient) list_command(){
	for channel_name, _ := range channels {
		users := channels[channel_name].chan_clients
		fmt.Fprintf(client.conn, ":" + hostname + " 322 " + client.nickname + " " + channel_name + " %d :\r\n", len(users));
	}
		fmt.Fprintf(client.conn, ":" + hostname + " 323 " + client.nickname + " :End of /LIST\r\n");
}


func (client *ConnectedClient) mode_command(args []string){
	date := channels[args[1]].created
	fmt.Fprintf(client.conn, ":" + hostname + " 329 " + client.nickname + " " + args[1] + " %d\r\n", date);
}

func (client *ConnectedClient) pass_command(args []string){
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
		if (args[1][0] == '#'){
			chan_lst := strings.Split(args[1], ",")
			for i := range chan_lst {
				client.join_channel(chan_lst[i])
				client.names2_command(chan_lst[i])
				}
		} else { client.respond_raw(":" + hostname + "NOTICE " + client.nickname + " :Erroneous channel name\r\n")
		}

	} else if fisrt_cmd == "WHO" {
		if (args[1][0] == '#'){
			client.who_command(args)
		}
	} else if fisrt_cmd == "MODE" {
		if (args[1][0] == '#'){
			client.mode_command(args)
		}
	} else if fisrt_cmd == "NAMES" {
		if (args[1][0] == '#') {
			client.names_command(args)
		} 
	} else if fisrt_cmd == "PART" {
		if (args[1][0] == '#'){
			client.part_command(args)
		}
	} else if fisrt_cmd == "LIST" {
			client.list_command()
	} else if fisrt_cmd == "PASS" {
		if (args[1] != ""){
			client.pass_command(args)
		}
	} else if fisrt_cmd == "CAP" {
	    client.cap_command(args)
	}  else {
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
		client.handle_command(msg)
	}

	client.unregister()
	conn.Close()
}

func main() {
	clients = make(map[string]*ConnectedClient)
	channels = make(map[string]*Channel)

	fmt.Println("Hello, 世界    " + hostname + ":" + port)
	listener, err := net.Listen("tcp", hostname + ":" + port)

    if (err == nil){
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error")
			continue
		}
		go handle_conn(conn)
	}
  } else {
  fmt.Println("error: could not open port ?")
  }
}

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"chat/userdb"
)

const (
	Who      = "/who"
	Nick     = "/nick"
	Register = "/register"
)

type Client struct {
	*bufio.Reader
	net.Conn
	server   *Server
	userName string
}

func NewClient(c net.Conn, s *Server) *Client {
	return &Client{
		Reader: bufio.NewReader(c),
		Conn:   c,
		server: s,
	}
}

func (c *Client) Send(str string, args ...interface{}) {
	c.Write([]byte(fmt.Sprintf(str+"\n", args...)))
}

type ClientMsg struct {
	client *Client
	msg    string
}

func (c *Client) Run() {
	for {
		msg, err := c.ReadString('\n')
		if err != nil {
			c.server.Leave <- c
			break
		}
		c.server.Msg <- &ClientMsg{
			client: c,
			msg:    msg[:len(msg)-1],
		}
	}
}

type Server struct {
	Join  chan *Client    // chan for new clients joining
	Leave chan *Client    // chan for clients leaving
	Msg   chan *ClientMsg // chan for clients posting msg

	activeUsers map[string]*Client
	userdb      userdb.UserDb
}

func NewServer(userdb userdb.UserDb) *Server {
	return &Server{
		Join:  make(chan *Client),
		Leave: make(chan *Client),
		Msg:   make(chan *ClientMsg),

		activeUsers: make(map[string]*Client),
		userdb:      userdb,
	}
}

func (s *Server) Broadcast(str string, args ...interface{}) {
	msg := fmt.Sprintf(str, args...)
	log.Println(msg)
	for _, c := range s.activeUsers {
		c.Send(msg)
	}
}

func (s *Server) MakeWhoResponse() string {
	if len(s.activeUsers) == 0 {
		return "No active users"
	}
	msg := "Active users: "
	first := true
	for u, _ := range s.activeUsers {
		if first {
			msg += u
			first = false
		} else {
			msg += fmt.Sprintf(", %v", u)
		}
	}
	return msg
}

func (s *Server) Run() {
	for {
		select {

		case c := <-s.Join:
			c.userName = s.userdb.MakeNewGuestName()
			c.Send("Welcome to the chat, %v!", c.userName)
			c.Send(s.MakeWhoResponse())

			s.activeUsers[c.userName] = c
			s.Broadcast("User %v joined", c.userName)
			go c.Run()

		case c := <-s.Leave:
			delete(s.activeUsers, c.userName)
			c.Close()
			s.Broadcast("User %v left", c.userName)

		case m := <-s.Msg:
			if len(m.msg) == 0 {
				continue
			}
			if m.msg[0] == '/' {
				parts := strings.Fields(m.msg)

				switch parts[0] {

				case Who:
					if len(parts) > 1 {
						m.client.Send("Too many arguments for %v command", Who)
						continue
					}
					m.client.Send(s.MakeWhoResponse())

				case Nick:
					if len(parts) < 2 || len(parts) > 3 {
						m.client.Send("Wrong number of arguments for %v command. Use '%v <nick>' to change nick or '%v <nick> <password>' to log in", Nick, Nick, Nick)
						continue
					}
					if len(parts) == 2 { // change nick
						name := parts[1]
						if _, exists := s.activeUsers[name]; exists {
							m.client.Send("Cant use nick %v: already in use", name)
							continue
						} else if s.userdb.IsRegistered(name) {
							m.client.Send("Nick %v is registered. Use '%v <nick> <password>' to log in", name, Nick)
							continue
						}
						s.Broadcast("User %v changed their nick to %v", m.client.userName, name)
						delete(s.activeUsers, m.client.userName)
						m.client.userName = name
						s.activeUsers[name] = m.client
					} else if len(parts) == 3 { // login
						name := parts[1]
						password := parts[2]
						if err := s.userdb.Authenticate(name, password); err != nil {
							m.client.Send(err.Error())
							continue
						} else if _, exists := s.activeUsers[name]; exists {
							s.activeUsers[name].Send("Someone took over your nick")
							s.activeUsers[name].Close()
						}
						s.Broadcast("User %v changed their nick to %v", m.client.userName, name)
						m.client.userName = name
						s.activeUsers[name] = m.client
					}

				case Register:
					if len(parts) != 2 {
						m.client.Send("Wrong number of arguments for %v command. Use '%v <nick> <password>' to register a nick", Register)
						continue
					}
					password := parts[1]
					if err := s.userdb.Register(m.client.userName, password); err != nil {
						m.client.Send(err.Error())
						continue
					}
					s.Broadcast("User %v registered", m.client.userName)

				default:
					m.client.Send("Unknown server command %v", parts[0])
					continue
				}
			} else {
				s.Broadcast("%v: %v", m.client.userName, m.msg)
			}
		}
	}
}

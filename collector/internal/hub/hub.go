package hub

// Client represents a connected WebSocket subscriber.
type Client struct {
	send   chan string
	filter string // "namespace/pod" or "" to receive all
}

func (c *Client) Send() <-chan string { return c.send }

type message struct {
	key  string
	line string
}

// Hub fans out log lines to all registered WebSocket clients.
// All map mutations happen in a single goroutine (Run) via channels — no mutex needed.
type Hub struct {
	reg   chan *Client
	unreg chan *Client
	bcast chan message
}

func New() *Hub {
	return &Hub{
		reg:   make(chan *Client, 64),
		unreg: make(chan *Client, 64),
		bcast: make(chan message, 4096),
	}
}

func (h *Hub) Run() {
	clients := make(map[*Client]struct{})
	for {
		select {
		case c := <-h.reg:
			clients[c] = struct{}{}

		case c := <-h.unreg:
			if _, ok := clients[c]; ok {
				delete(clients, c)
				close(c.send)
			}

		case msg := <-h.bcast:
			for c := range clients {
				if c.filter != "" && c.filter != msg.key {
					continue
				}
				select {
				case c.send <- msg.line:
				default:
					// slow consumer: evict to prevent back-pressure
					delete(clients, c)
					close(c.send)
				}
			}
		}
	}
}

// Register returns a new client subscribed to the given pod key ("ns/pod") or all pods ("").
func (h *Hub) Register(filter string) *Client {
	c := &Client{send: make(chan string, 256), filter: filter}
	h.reg <- c
	return c
}

func (h *Hub) Unregister(c *Client) {
	h.unreg <- c
}

func (h *Hub) Broadcast(key, line string) {
	select {
	case h.bcast <- message{key: key, line: line}:
	default:
		// drop when broadcast buffer is full (log storm protection)
	}
}

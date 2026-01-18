# Galliard

**Galliard** is a modular, idiomatic, and extensible [Bayeux protocol](https://docs.cometd.org/current/reference/#_bayeux) server library for Go.  
It’s designed for building real-time, channel-based publish/subscribe systems—think chat, live dashboards, or event streaming.  
Inspired by Faye and CometD, Galliard aims to be simple, robust, and easy to extend.

---

## Features

- **Bayeux protocol**: handshake, connect, subscribe, unsubscribe, publish, disconnect
- **Channel-based pub/sub**: subscribe and publish to any channel
- **Thread-safe**: fine-grained locking for high concurrency
- **Minimal public API**: just what you need, nothing you don’t
- **Extensible**: ready for server-side hooks, custom channels, and more

---

## Implementation Checklist

Here’s what’s ready and what’s coming soon:

- [x] Bayeux protocol core (handshake, connect, subscribe, unsubscribe, publish, disconnect)
- [x] Channel-based pub/sub
- [x] Thread-safe session and channel management
- [x] Per-session advice and protocol-compliant error handling
- [x] Minimal, clean public API (`Server`, `NewServer`, `HandleMessage`)
- [x] GoDoc comments and usage examples
- [ ] Server-side event hooks (`Subscribe`, `OnDisconnect`, etc.)
- [ ] Go Bayeux client package
- [ ] HTTP/WebSocket transport helpers
- [ ] More real-world examples and advanced documentation

*Want to help or need a feature? Open an issue or PR!*

---

## Getting Started

### 1. Install

```bash
go get github.com/charlinchui/galliard
```

### 2. Usage Example

``` go
import (
    "github.com/charlinchui/galliard/server"
    "github.com/charlinchui/galliard/message"
    "fmt"
)

func main() {
    srv := server.NewServer()

    // Example: Handle a handshake
    req := &message.BayeuxMessage{Channel: "/meta/handshake"}
    resp := srv.HandleMessage(req)
    fmt.Printf("Handshake response: %+v\n", resp)
}
```

### 3. Typical Bayeux Flow

- **Handshake:**  
  Client sends `/meta/handshake`, receives a `clientId`.
- **Subscribe:**  
  Client sends `/meta/subscribe` with `clientId` and channel.
- **Publish:**  
  Client sends a message to a channel.
- **Connect:**  
  Client sends `/meta/connect` to receive messages (long-polling or WebSocket).
- **Unsubscribe/Disconnect:**  
  Client can unsubscribe or disconnect gracefully.

---

## Project Structure

``` text
galliard/
  server/      # Bayeux server implementation (public API)
  message/     # Bayeux message and advice types (public API)
  internal/    # Internal packages (client, channel, utils)
``` 
---

## Public API

- `type Server`  
  The Bayeux server.
- `func NewServer() *Server`  
  Create a new server.
- `func (s *Server) HandleMessage(msg *message.BayeuxMessage) *message.BayeuxMessage`  
  Process a Bayeux message and get a response.
- `type BayeuxMessage`  
  The protocol message type (in `message` package).
- `type Advice`  
  Connection advice for clients (in `message` package).

---

## Extending Galliard

- **Server-side event hooks:**  
  (e.g., `Subscribe`, `OnDisconnect`) are planned for Go-side event handling.
- **Custom channels and business logic:**  
  Can be implemented by extending the server or adding hooks.
- **Client package:**  
  A Go Bayeux client is on the roadmap.

---

## Running Tests

``` bash
go test ./...
```

---

## Contributing

Galliard is open to contributions!  
If you have ideas, bug reports, or want to help with features, please open an issue or pull request.  
Let’s make real-time Go apps easy and fun together.

---

## License

MIT License

---

## References

- [Bayeux Protocol Specification](https://docs.cometd.org/current/reference/#_bayeux)
- [CometD Project](https://cometd.org/)
- [Faye Project](https://faye.jcoglan.com/)

---

**Galliard**: Real-time, reliable, and readable pub/sub for Go.

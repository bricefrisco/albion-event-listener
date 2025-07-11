# Albion Event Listener
Albion event listener written in Golang.

Most code coming from [albiondata-client](https://github.com/ao-data/albiondata-client), with minor adjustments.

Intended to be used as a minimal dependency to build other things.

---

Usage:
```go
msgChan := make(chan *Message, 1000)
l := listener.NewListener(msgChan)
go l.Run()

for message := range msgChan {
  if message.Type === "event" {
    // Do something with event
    fmt.Println(message)
  } else if message.Type === "operation" { 
    // Do something with operation
    fmt.Println(message)
  } else { 
    // Do something with movement event 
    fmt.Println(message)
  }
}
```
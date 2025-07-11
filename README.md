# Albion Listener
Albion event and operation listener written in Golang.

Most code coming from [albiondata-client](https://github.com/ao-data/albiondata-client), with minor adjustments.

Intended to be used as a minimal dependency to build other things.

---

Usage:
```go
msgChan := make(chan *Message, 1000)
l := listener.NewListener(msgChan)
go l.Run()

for message := range msgChan {
    // Do something with the message
    fmt.Println(message.Type, message.Name, message.Data)
  }
}
```
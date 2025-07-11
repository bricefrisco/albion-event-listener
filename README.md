# Albion Event Listener
Albion event listener written in Golang.

Most code coming from [albiondata-client](https://github.com/ao-data/albiondata-client), with minor adjustments.

Intended to be used as a minimal dependency to build other things.

---

Usage:
```go
msgChan := make(chan map[uint8]any, 1000)
listener := photon.NewListener(msgChan)
go listener.Run()

for message := range msgChan {
  if message[252] != nil {
    // Do something with event
    fmt.Println("event", message)
  } else if message[253] != nil { 
    // Do something with operation
    fmt.Println("operation", message)
  } else { 
    // Do something with movement event 
    fmt.Println("movement", message)
  }
}
```
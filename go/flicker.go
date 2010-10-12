/**
 * Code forked from http://gist.github.com/505923 under the Apache 2 license.
 *
 * neildunn@gmail.com
 */

package main

import (
    "flag"
    "http"
    "log"
    "template"
    "websocket"
    "time"
    "fmt"
    "bytes"
    "os"
)

var addr = flag.String("addr", ":8080", "http service address")
var updateIntervalMs = flag.Int64("updateIntervalMs", 100, "update interval of Websocket in MS")
var numWindows = flag.Int("numWindows", 10, "number of windows to run the benchmark in")
var padding = flag.Int("padding", 0, "padding to be added to the message (to increase WS message size)")
var width = flag.Int("width", 200, "width of windows")
var height = flag.Int("height", 800, "height of windows")

func main() {
    flag.Parse()

    var hostname, _ = os.Hostname()
    
    fmt.Printf("Running at: http://%s:%s/\n", hostname, *addr)

    var ticker = time.Tick(*updateIntervalMs * 1000 * 1000);
    go hub(ticker)
    
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/ws", webSocketProtocolSwitch)

    if err := http.ListenAndServe(*addr, nil); err != nil {
        log.Exit("ListenAndServe:", err)
    }
}

func webSocketProtocolSwitch(c http.ResponseWriter, req *http.Request) {
    // Handle old and new versions of protocol.
    if _, found := req.Header["Sec-Websocket-Key1"]; found {
        websocket.Handler(clientHandler).ServeHTTP(c, req)
    } else {
        websocket.Draft75Handler(clientHandler).ServeHTTP(c, req)
    }
}

var messageChan = make(chan []byte)

type subscription struct {
    conn *websocket.Conn
    subscribe bool
}

var subscriptionChan = make(chan subscription)

func hub(tickerChan <-chan int64) {
    conns := make(map[*websocket.Conn]int)
    for {
        select {
        case subscription := <-subscriptionChan:
            conns[subscription.conn] = 0, subscription.subscribe
	case tick := <-tickerChan:
	    for conn, _ := range conns {
 		var message = bytes.NewBufferString(fmt.Sprintf("%d,", tick));

		for i := 0; i < *padding; i++ {
 		   message.WriteString("A");
		}	
	
                if _, err := conn.Write(message.Bytes()); err != nil {
                    conn.Close()
                }
            }
	}
    }
}

func clientHandler(ws *websocket.Conn) {
    defer func() {
        subscriptionChan <- subscription{ws, false}
        ws.Close()
    }()

    subscriptionChan <- subscription{ws, true}

    buf := make([]byte, 256)
    for {
        n, err := ws.Read(buf)
        if err != nil {
            break
        }
        messageChan <- buf[0:n]
    }
}

type templateData struct {
    host string
    numWindows int
    width int
    height int
}

// Handle home page requests.
func homeHandler(c http.ResponseWriter, req *http.Request) {
    homeTempl.Execute(templateData{req.Host, *numWindows, *width, *height}, c)
}

var homeTempl *template.Template

func init() {
    homeTempl = template.New(nil)
    homeTempl.SetDelims("«", "»")
    if err := homeTempl.Parse(homeStr2); err != nil {
        panic("template error: " + err.String())
    }
}

const homeStr2 = `
<html>
<head>
<style type="text/css">
body {
  color: white;
  background-color: black;
}

.block {
  width: 200px;
  height: 20px;
  margin-bottom: 5px;
  background-color: #ccc;
  color: black;
}
</style>
<script>
conn = new WebSocket("ws://«host»/ws");
conn.onclose = function(evt) {
   document.body.innerHTML += "<p>Connection closed!</p>"; 
}

conn.onopen = function(evt) {
   document.body.innerHTML += "<p>Connection open!</p>"; 
}

conn.onmessage = function(evt) {
   var blocks = document.getElementsByClassName("block");
   var n = Math.floor(Math.random() * (blocks.length));
   var block = blocks[n];
   block.innerHTML = "<p>" + evt.data.split(",")[0] + "</p>";
}

function launch() {
  for (var i = 0; i < «numWindows»; i++) {
    var result = window.open("http://«host»", null, "dialog=true,width=«width»px,height=«height»px");
    result.moveTo((i * «width») - (Math.floor(i / 10) * (10 * «height»)), Math.floor(i / 10) * «height»);
  }
}
</script>
</head>
<body>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<div class="block"></div>
<button onclick="launch()">Launch</button>
</body>
</html> `


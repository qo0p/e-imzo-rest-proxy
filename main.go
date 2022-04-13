package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	apiPort        = flag.Int("p", 64647, "REST-API listen port")
	logLevel       = flag.Int("log-level", int(logrus.InfoLevel), "Log level (0-panic, 1-fatal, 2-error, 3-warn, 4-info, 5-debug, 6-trace, 7-dump, 8-callreport)")
	wsURL          = url.URL{Scheme: "ws", Host: "127.0.0.1:64646", Path: "/service/cryptapi"}
	lock           = new(sync.Mutex)
	readTimeout    = 30 * time.Second
	connectTimeout = 5 * time.Second
	hdr            = make(http.Header)
)

func main() {
	flag.Parse()

	hdr.Set("Origin", "http://localhost")

	logrus.SetLevel(logrus.Level(*logLevel))
	mux := http.NewServeMux()
	mux.HandleFunc("/", proxyfy)
	logrus.Infof("rest-api is listening on localhost:%v", *apiPort)
	err := http.ListenAndServe(fmt.Sprintf("localhost:%v", *apiPort), mux)
	if err != nil {
		logrus.Fatal(err)
	}

}

func proxyfy(w http.ResponseWriter, r *http.Request) {

	lock.Lock()
	defer lock.Unlock()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	reply, err := eimzoIO(body)
	if err != nil {
		logrus.Errorf("error io: %v", err)
		http.Error(w, "can't io with e-imzo", http.StatusBadGateway)
		return
	}
	if reply == nil {
		http.Error(w, "no reply from e-imzo", http.StatusGatewayTimeout)
		return
	}
	w.Write(reply)
}

func eimzoIO(req []byte) ([]byte, error) {
	logrus.Printf("connecting to %s", wsURL.String())
	ctx, _ := context.WithTimeout(context.Background(), connectTimeout)
	c, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), hdr)
	if err != nil {
		logrus.Errorf("dial: %v", err)
		return nil, err
	}
	defer func() {
		c.Close()
		logrus.Printf("disconnected from %s", wsURL.String())
	}()

	reply := make(chan []byte)

	go func() {
		defer close(reply)
		_, message, err := c.ReadMessage()
		if err != nil {
			logrus.Errorf("read: %v", err)
			return
		}
		if logrus.GetLevel() == logrus.TraceLevel {
			fmt.Printf("recv json:\n%v\n", string(message))
		}
		reply <- message
	}()

	defer func() {
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			logrus.Errorf("write close: %v", err)
			return
		}
	}()
	if logrus.GetLevel() == logrus.TraceLevel {
		fmt.Printf("sent json:\n%v\n", string(req))
	}
	err = c.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		logrus.Errorf("write: %v", err)
		return nil, err
	}

	select {
	case m := <-reply:
		return m, nil
	case <-time.After(readTimeout):
		return nil, nil
	}
}

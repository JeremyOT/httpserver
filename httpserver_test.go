package httpserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestServeStop(t *testing.T) {
	requests := map[string]*http.Request{}
	shutdownChannel := make(chan struct{})
	server := New(func(writer http.ResponseWriter, request *http.Request) {
		requests[request.URL.Path] = request
		writer.Write([]byte("OK"))
	})
	if err := server.Start("---"); err == nil {
		t.Fatal("Expected failure on bad address")
	}
	server.Start("")
	server.SetShutdownHandler(func() {
		close(shutdownChannel)
	})
	<-server.WaitForStart()
	addr := server.Address()
	response, err := http.Get(fmt.Sprintf("http://%s/test-path", addr))
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if string(body) != "OK" {
		t.Fatalf("Expected \"OK\" received \"%s\"", string(body))
	}
	if _, ok := requests["/test-path"]; !ok {
		t.Fatal("Request not received")
	}
	select {
	case <-shutdownChannel:
		t.Fatal("Shutdown channel closed prematurely")
	case <-time.After(time.Microsecond):
	}
	<-server.Stop()
	select {
	case <-shutdownChannel:
	case <-time.After(time.Microsecond):
		t.Fatal("Shutdown not closed after shutdown")
	}
	_, err = http.Get(fmt.Sprintf("http://%s/closed", addr))
	if err == nil {
		t.Fatal("Expected error:", err)
	}
	if _, ok := requests["/closed"]; ok {
		t.Fatal("Request received after shutdown")
	}
}

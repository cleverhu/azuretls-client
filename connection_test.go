package azuretls

import (
	"runtime"
	"sync"
	"testing"
)

func TestSessionConn(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	session := NewSession()

	response, err := session.Get("https://httpbin.org/get")

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != 200 {
		t.Fatal("TestHeader failed, expected: 200, got: ", response.StatusCode)
	}

	if len(session.conns.hosts) == 0 {
		t.Fatal("TestSessionConn failed, Conn is empty")
	}

	firstConn := session.conns.hosts["httpbin.org:443"]

	if !firstConn.http2Conn.CanTakeNewRequest() {
		t.Fatal("TestSessionConn failed, Conn is not reusable")
	}

	if err = firstConn.tlsConn.VerifyHostname("httpbin.org"); err != nil {
		t.Fatal("TestSessionConn failed, VerifyHostname failed : ", err)
	}

	response, err = session.Get("https://httpbin.org/get")

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != 200 {
		t.Fatal("TestHeader failed, expected: 200, got: ", response.StatusCode)
	}

	if len(session.conns.hosts) != 1 {
		t.Fatal("TestSessionConn failed, Conn is not reused")
	}

	if firstConn != session.conns.hosts["httpbin.org:443"] {
		t.Fatal("TestSessionConn failed, Conn is not reused")
	}
}

func TestHTTP1Conn(t *testing.T) {
	session := NewSession()

	_, err := session.Get("https://api.ipify.org/")

	if err != nil {
		t.Fatal(err)
	}
}

func TestHighConcurrency(t *testing.T) {

	session := NewSession()

	wait := &sync.WaitGroup{}
	wait.Add(1000)

	var err error
	var ok int

	for i := 0; i < 1000; i++ {
		go func() {
			defer wait.Done()
			_, err2 := session.Get("https://example.com")

			if err2 != nil {
				err = err2
				t.Error(ok)
				t.Fatal(err2)
				return
			}

			ok++
		}()
	}

	wait.Wait()

	if err != nil {
		t.Fatal(err)
	}

	if ok != 1000 {
		t.Fatal("TestHighConcurrency failed, expected: 1000, got: ", ok)
	}
}

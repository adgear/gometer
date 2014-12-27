// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCarbonHandler(t *testing.T) {

	c0 := &TestCarbon{T: t, Name: "c0"}
	c1 := &TestCarbon{T: t, Name: "c1"}

	c0.Init()
	c1.Init()

	time.Sleep(100 * time.Millisecond)

	CarbonDialTimeout = 10 * time.Millisecond
	CarbonMaxConnDelay = 100 * time.Millisecond
	handler := &CarbonHandler{URLs: []string{c0.URL, c1.URL}}
	handler.Init()

	CarbonSend("init", handler, c0, c1)

	c0.Stop()
	CarbonSend("stop-c0", handler, c1)

	c1.Stop()
	CarbonSend("stop-c1", handler)

	// Should not be necessary but, for some random reason, the closed
	// connection is only detected on the second round of sending.
	CarbonSend("stop-c1-2", handler)

	c1.Start()
	CarbonSend("start-c1", handler, c1)

	c0.Start()
	CarbonSend("start-c0", handler, c0, c1)
}

func CarbonSend(title string, handler *CarbonHandler, carbons ...*TestCarbon) {
	values := map[string]float64{"a": 1, "b": 2}
	time.Sleep(200 * time.Millisecond)

	handler.HandleMeters(values)
	for _, carbon := range carbons {
		carbon.Expect(title, values)
	}
}

type CarbonPair struct {
	Key   string
	Value float64
}

type TestCarbon struct {
	T    *testing.T
	Name string
	URL  string

	initialize sync.Once

	pairC chan CarbonPair

	stopC  chan int
	startC chan int

	listenC chan net.Listener
	connC   chan net.Conn

	listen net.Listener
	conns  []net.Conn
}

func (carbon *TestCarbon) Init() {
	carbon.initialize.Do(carbon.init)
}

func (carbon *TestCarbon) Start() {
	carbon.Init()
	carbon.startC <- 1
}

func (carbon *TestCarbon) Stop() {
	carbon.Init()
	carbon.stopC <- 1
}

func (carbon *TestCarbon) Expect(title string, exp map[string]float64) {
	carbon.Init()

	values := make(map[string]float64)

	done := false
	timeoutC := time.After(100 * time.Millisecond)

	for !done {
		select {
		case pair := <-carbon.pairC:
			values[pair.Key] = pair.Value
			done = len(values) == len(exp)

		case <-timeoutC:
			done = true
		}
	}

	CheckValues(carbon.T, fmt.Sprintf("%s.%s", title, carbon.Name), values, exp)
}

func (carbon *TestCarbon) init() {
	carbon.pairC = make(chan CarbonPair, 1<<16)
	carbon.stopC = make(chan int)
	carbon.startC = make(chan int)
	carbon.listenC = make(chan net.Listener)
	carbon.connC = make(chan net.Conn)

	go carbon.accept()

	listen := <-carbon.listenC
	fmt.Printf("LOG(%s): listening on: %s\n", carbon.Name, listen.Addr().String())
	carbon.URL = listen.Addr().String()
	carbon.listen = listen

	go carbon.run()
}

func (carbon *TestCarbon) accept() {
	for {
		URL := "127.0.0.1:0"
		if carbon.URL != "" {
			URL = carbon.URL
		}

		listen, err := net.Listen("tcp", URL)
		if err != nil {
			carbon.T.Fatalf("FATAL(%s): unable to listen: %s", carbon.Name, err)
			return
		}

		carbon.listenC <- listen

		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Printf("LOG(%s): accept failed: %s\n", carbon.Name, err)
				break
			}

			carbon.connC <- conn
			go carbon.handle(conn)
		}

		<-carbon.startC
		fmt.Printf("LOG(%s): started\n", carbon.Name)
	}
}

func (carbon *TestCarbon) handle(conn net.Conn) {
	fmt.Printf("LOG(%s): new connection\n", carbon.Name)

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("LOG(%s): read failed: %s\n", carbon.Name, err)
			return
		}

		split := strings.Split(line[:len(line)-1], " ")

		if len(split) != 3 {
			carbon.T.Errorf("FAIL(%s): not enough pieces in '%s' from '%s'",
				carbon.Name, split, line)
			continue
		}

		if _, err := strconv.ParseInt(split[2], 10, 64); err != nil {
			carbon.T.Errorf("FAIL(%s): parse failed '%s' in '%s' -> %s",
				carbon.Name, split[2], line, err)
		}

		value, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			carbon.T.Errorf("FAIL(%s): parse failed '%s' in '%s' -> %s",
				carbon.Name, split[1], line, err)
			continue
		}

		fmt.Printf("LOG(%s): pair {%s, %f}\n", carbon.Name, split[0], value)
		carbon.pairC <- CarbonPair{split[0], value}
	}
}

func (carbon *TestCarbon) run() {
	for {
		select {
		case listen := <-carbon.listenC:
			fmt.Printf("LOG(%s): listening on: %s\n", carbon.Name, listen.Addr().String())
			carbon.URL = listen.Addr().String()
			carbon.listen = listen

		case conn := <-carbon.connC:
			carbon.conns = append(carbon.conns, conn)

		case <-carbon.stopC:
			carbon.listen.Close()
			carbon.listen = nil

			for _, conn := range carbon.conns {
				conn.Close()
			}
			carbon.conns = nil
			fmt.Printf("LOG(%s): stopped\n", carbon.Name)
		}
	}
}

package server

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/xesina/message-delivery/internal/client"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

const testAddr = "localhost:50005"

type connMock struct {
	net.Conn
}

type ServerTestSuite struct {
	suite.Suite
	server *Server
}

func (suite *ServerTestSuite) SetupSuite() {
	suite.server = New()
	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	if err != nil {
		suite.FailNow("resolving address failed %s", err)
	}

	go func() {
		err := suite.server.Start(tcpAddr)
		if err != nil {
			fmt.Println("starting server failed: ", err)
			os.Exit(1)
		}
	}()

}

func (suite *ServerTestSuite) TearDownSuite() {
	//suite.server.Stop()
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) clientsCount() int {
	suite.server.cl.RLock()
	defer suite.server.cl.RUnlock()
	return len(suite.server.clients)
}

func (suite *ServerTestSuite) handlersCount() int {
	suite.server.hl.RLock()
	defer suite.server.hl.RUnlock()
	return len(suite.server.handler)
}

func (suite *ServerTestSuite) resetIdCounter() {
	atomic.StoreUint64(&suite.server.id, 0)
}

func (suite *ServerTestSuite) TestRegisterHandlers() {
	l := suite.handlersCount()

	// to ensure each time we register handler we update the test to
	// control the registration of the handlers
	suite.Equal(3, l)
}

func (suite *ServerTestSuite) TestRegisterClient() {
	type connMock struct {
		net.Conn
	}

	tt := []struct {
		conn net.Conn
	}{
		{conn: &connMock{}},
		{conn: &connMock{}},
		{conn: &connMock{}},
	}

	for c, tc := range tt {
		suite.server.registerClient(tc.conn)
		suite.Equal(c+1, suite.clientsCount())
	}
}

func (suite *ServerTestSuite) TestDeregisterClients() {
	tt := []struct {
		conn net.Conn
	}{
		{conn: &connMock{}},
		{conn: &connMock{}},
		{conn: &connMock{}},
	}

	registerAll := func(conns []struct{ conn net.Conn }) {
		for _, c := range conns {
			suite.server.registerClient(c.conn)
		}
	}

	deregisterAll := func(conns []struct{ conn net.Conn }) {
		for _, c := range conns {
			suite.server.deregisterClient(c.conn)
		}
	}

	// test deregister right after a register
	for _, tc := range tt {
		suite.server.registerClient(tc.conn)
		suite.server.deregisterClient(tc.conn)
		suite.Equal(0, suite.clientsCount())
	}

	// first register all the clients and then deregister all
	registerAll(tt)
	deregisterAll(tt)
	suite.Equal(0, suite.clientsCount())
}

func (suite *ServerTestSuite) TestStart() {
	conn, err := net.Dial("tcp", testAddr)
	defer conn.Close()
	suite.NoError(err)
}

func (suite *ServerTestSuite) TestHandleFunc() {
	l := suite.handlersCount()

	msgName := "testHandler"
	testHandler := func(ctx *context) error { return nil }
	suite.server.HandleFunc(msgName, testHandler)

	newLen := suite.handlersCount()
	suite.Equal(l+1, newLen)

	suite.server.hl.RLock()
	_, ok := suite.server.handler[msgName]
	suite.True(ok)
	suite.server.hl.RUnlock()

	suite.server.hl.Lock()
	delete(suite.server.handler, msgName)
	suite.server.hl.Unlock()
}

func (suite *ServerTestSuite) TestHandleIdentity() {
	suite.resetIdCounter()

	cl := client.New()

	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)

	err = cl.Connect(tcpAddr)
	defer cl.Close()
	suite.NoError(err)

	id, err := cl.WhoAmI()
	suite.NoError(err)
	suite.Equal(uint64(1), id)
}

func (suite *ServerTestSuite) TestListClientIDs() {
	tt := []struct {
		conns []*connMock
		want  []uint64
	}{
		{
			conns: []*connMock{
				{},
			},
			want: []uint64{1},
		},
		{
			conns: []*connMock{},
			want:  []uint64{},
		},
		{
			conns: []*connMock{
				{},
				{},
				{},
				{},
			},
			want: []uint64{1, 2, 3, 4},
		},
	}

	for _, tc := range tt {
		suite.resetIdCounter()

		for _, conn := range tc.conns {
			suite.server.registerClient(conn)
		}

		ids := suite.server.ListClientIDs()
		suite.Equal(len(tc.want), len(ids))
		suite.ElementsMatch(tc.want, ids)

		for _, conn := range tc.conns {
			suite.server.deregisterClient(conn)
		}
	}
}

func (suite *ServerTestSuite) TestHandleListWithSingleClient() {
	suite.resetIdCounter()

	cl := client.New()

	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)

	err = cl.Connect(tcpAddr)
	defer cl.Close()
	suite.NoError(err)

	ids, err := cl.ListClientIDs()
	suite.NoError(err)
	suite.ElementsMatch([]uint64{}, ids)
}

func (suite *ServerTestSuite) TestHandleListWithMultipleClient() {
	suite.resetIdCounter()

	cl1 := client.New()
	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl1.Connect(tcpAddr)
	defer cl1.Close()
	suite.NoError(err)

	cl2 := client.New()
	tcpAddr, err = net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl2.Connect(tcpAddr)
	defer cl2.Close()
	suite.NoError(err)

	cl3 := client.New()
	tcpAddr, err = net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	// intentionally we won't defer this client's Close() and we
	// manually close to test another case
	err = cl3.Connect(tcpAddr)
	suite.NoError(err)

	time.Sleep(100 * time.Millisecond)

	// checking with client2 so should get ids: 1,3
	expedtedIds := []uint64{1, 3}
	ids, err := cl2.ListClientIDs()
	suite.NoError(err)
	suite.Len(ids, len(expedtedIds))
	suite.ElementsMatch(expedtedIds, ids)

	// closing client3 and checking with client1 to get only id:2
	cl3.Close()
	expedtedIds = []uint64{2}
	ids, err = cl1.ListClientIDs()
	suite.NoError(err)
	suite.Len(ids, len(expedtedIds))
	suite.ElementsMatch(expedtedIds, ids)
}

func (suite *ServerTestSuite) TestSendToOneClient() {
	suite.resetIdCounter()

	cl1 := client.New()
	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl1.Connect(tcpAddr)
	defer cl1.Close()
	suite.NoError(err)

	cl1Ch := make(chan client.IncomingMessage)
	go cl1.HandleIncomingMessages(cl1Ch)

	cl2 := client.New()
	tcpAddr, err = net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl2.Connect(tcpAddr)
	defer cl2.Close()
	suite.NoError(err)

	cl2Ch := make(chan client.IncomingMessage)
	go cl2.HandleIncomingMessages(cl2Ch)

	time.Sleep(100 * time.Millisecond)

	receiver := uint64(1)
	expectedSenderId := uint64(2)
	expectedBody := "Hello"

	err = cl2.SendMsg([]uint64{receiver}, []byte(expectedBody))
	suite.NoError(err)
	incomingFromCl2 := <-cl1Ch
	suite.Equal(incomingFromCl2.SenderID, expectedSenderId)
	suite.Equal(string(incomingFromCl2.Body), expectedBody)
}

func (suite *ServerTestSuite) TestSendToTwoClient() {
	suite.resetIdCounter()

	cl1 := client.New()
	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl1.Connect(tcpAddr)
	defer cl1.Close()
	suite.NoError(err)

	cl1Ch := make(chan client.IncomingMessage)
	go cl1.HandleIncomingMessages(cl1Ch)

	cl2 := client.New()
	tcpAddr, err = net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl2.Connect(tcpAddr)
	defer cl2.Close()
	suite.NoError(err)

	cl2Ch := make(chan client.IncomingMessage)
	go cl2.HandleIncomingMessages(cl2Ch)

	cl3 := client.New()
	tcpAddr, err = net.ResolveTCPAddr("tcp", testAddr)
	suite.NoError(err)
	err = cl3.Connect(tcpAddr)
	defer cl3.Close()
	suite.NoError(err)

	cl3Ch := make(chan client.IncomingMessage)
	go cl3.HandleIncomingMessages(cl3Ch)

	time.Sleep(100 * time.Millisecond)

	receivers := []uint64{1, 2}
	expectedSenderId := uint64(3)
	expectedBody := "Hello"

	err = cl3.SendMsg(receivers, []byte(expectedBody))
	suite.NoError(err)

	incomingFromCl3 := <-cl1Ch
	suite.Equal(incomingFromCl3.SenderID, expectedSenderId)
	suite.Equal(string(incomingFromCl3.Body), expectedBody)

	incomingFromCl2 := <-cl2Ch
	suite.Equal(incomingFromCl2.SenderID, expectedSenderId)
	suite.Equal(string(incomingFromCl2.Body), expectedBody)
}

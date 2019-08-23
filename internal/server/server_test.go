package server

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/xesina/message-delivery/internal/client"
	"net"
	"os"
	"sync/atomic"
	"testing"
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
	suite.Equal(1, l)
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
	testHandler := func(conn net.Conn) error { return nil }
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

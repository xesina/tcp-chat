package server

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"net"
	"os"
	"testing"
)

const testAddr = "localhost:50005"

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

func (suite *ServerTestSuite) TearDownTest() {
	suite.server.cl.Lock()
	suite.server.clients = make(map[net.Conn]uint64)
	suite.server.cl.Unlock()
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

func (suite *ServerTestSuite) TestRegisterClient() {
	fmt.Println("running TestRegisterMultipleClient")

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

func (suite *ServerTestSuite) TestDeregisterClient() {
	fmt.Println("running TestRegisterMultipleClient")

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

func (suite *ServerTestSuite) TestRegisterHandlers() {
	suite.server.hl.RLock()
	l := len(suite.server.handler)
	suite.server.hl.RUnlock()

	// to ensure each time we register handler we update the test to
	// control the registration of the handlers
	suite.Equal(1, l)
}

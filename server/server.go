package server

import (
	"cs598fts/common"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Object struct {
	Lock      sync.Mutex
	Val       string
	Timestamp int
	ClientID  int
}

type Server struct {
	Addr      string
	ObjectMap sync.Map
}

func NewServer(addr string) (*Server, error) {
	server := &Server{Addr: addr}
	return server, nil
}

func (s *Server) Serve() error {
	// Setup rpc service
	if err := rpc.Register(s); err != nil {
		return err
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", s.Addr)
	if e != nil {
		return e
	}
	// Listen on addr
	runningTime := time.Now().String()
	logrus.Infof("Server is listening on %v and started at time: %v", s.Addr, runningTime)
	return http.Serve(l, nil)
}

func (s *Server) getObject(key string) *Object {
	newObject := &Object{
		Lock:      sync.Mutex{},
		Val:       "",
		Timestamp: 0,
		ClientID:  -1,
	}
	object, _ := s.ObjectMap.LoadOrStore(key, newObject)
	return object.(*Object)
}

func (s *Server) Set(req *common.SetRequest, resp *common.SetResponse) error {
	//logrus.Infof("Set request: %+v", req)
	// Grab object with key
	object := s.getObject(req.Key)
	object.Lock.Lock()
	defer object.Lock.Unlock()
	// Check if the req is newer
	if req.Timestamp > object.Timestamp || (req.Timestamp == object.Timestamp && req.ClientID > object.ClientID) {
		object.Timestamp = req.Timestamp
		object.ClientID = req.ClientID
		object.Val = req.Val

	}
	resp.Ok = true
	return nil
}

func (s *Server) Get(req *common.GetRequest, resp *common.GetResponse) error {
	//logrus.Infof("Get request: %+v", req)
	// Grab object with key
	object := s.getObject(req.Key)
	object.Lock.Lock()
	defer object.Lock.Unlock()
	// Check if key doesn't exist
	resp.Exist = object.ClientID != -1
	resp.Timestamp = object.Timestamp
	resp.Val = object.Val

	return nil
}

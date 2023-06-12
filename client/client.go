package client

import (
	"cs598fts/common"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/rpc"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Server []string `json:"server"`
}

type Client struct {
	server []*rpc.Client
	config *Config
	ID     int
}

func NewClient(clientID int, configPath string) (*Client, error) {
	// Setup client object
	client := &Client{server: make([]*rpc.Client, 0), ID: clientID}
	// Parse config
	config := &Config{}
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(configFile)
	if err := json.Unmarshal(byteValue, config); err != nil {
		return nil, err
	}
	client.config = config
	// Establish connection to servers
	for _, addr := range config.Server {
		rpcClient, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			return nil, err
		}
		client.server = append(client.server, rpcClient)
	}
	return client, nil
}

func (c *Client) majority() int {
	return (len(c.server) + 1) / 2
}

func (c *Client) get(key string) (bool, string, int, error) {
	getReq := &common.GetRequest{
		Key: key,
	}

	// Send getReq to servers and wait for response from majority
	var wg sync.WaitGroup
	wg.Add(len(c.server))
	var result sync.Map
	for idx, s := range c.server {
		go func(idx int, s *rpc.Client) {
			getResp := &common.GetResponse{}
			if err := s.Call("Server.Get", getReq, getResp); err != nil {
				logrus.Debugf("Server %s get err: %s", c.config.Server[idx], err)
			} else {
				result.Store(idx, getResp)
				//logrus.Debugf("Server %s get resp: %+v", c.config.Server[idx], getResp)
			}
			wg.Done()
		}(idx, s)
	}
	wg.Wait()
	// Get the largest timestamp
	serverCnt := 0
	existCnt := 0
	maxTimestamp := 0
	val := ""
	result.Range(func(k, v any) bool {
		serverCnt += 1
		resp := v.(*common.GetResponse)
		if !resp.Exist {
			return true
		}
		existCnt += 1
		if resp.Timestamp > maxTimestamp {
			maxTimestamp = resp.Timestamp
			val = resp.Val
		}
		return true
	})
	// Check majority
	if serverCnt < c.majority() {
		return false, "", 0, errors.New("not getting response from majority of servers")
	}

	return existCnt >= c.majority(), val, maxTimestamp, nil
}

func (c *Client) set(key string, val string, timestamp int) error {
	setReq := &common.SetRequest{
		Key:       key,
		Val:       val,
		Timestamp: timestamp,
		ClientID:  c.ID,
	}

	// Send setReq to servers and wait for response from majority
	var wg sync.WaitGroup
	wg.Add(len(c.server))
	var result sync.Map
	for idx, s := range c.server {
		go func(idx int, s *rpc.Client) {
			setResp := &common.SetResponse{}
			if err := s.Call("Server.Set", setReq, setResp); err != nil {
				logrus.Infof("Server %s set err: %s", c.config.Server[idx], err)
			} else {
				result.Store(idx, setResp)
				//logrus.Infof("Server %s set resp: %+v", c.config.Server[idx], setResp)
			}
			wg.Done()
		}(idx, s)
	}
	wg.Wait()

	// Get the largest timestamp
	serverCnt := 0
	okCnt := 0
	result.Range(func(k, v any) bool {
		serverCnt += 1
		resp := v.(*common.SetResponse)
		if resp.Ok {
			okCnt += 1
		}
		return true
	})
	// Check majority
	if serverCnt < c.majority() {
		return errors.New("not getting response from majority of servers")
	}

	return nil
}

func (c *Client) Write(key string, val string) error {
	// Get max-ts with val
	_, _, timestamp, err := c.get(key)
	if err != nil {
		return err
	}
	timestamp += 1
	// Ask storage to store
	// Note here the key is from user input
	if err := c.set(key, val, timestamp); err != nil {
		return err
	}
	return nil
}

func (c *Client) Read(key string) (string, error) {
	// Get max-ts with val
	exist, val, timestamp, err := c.get(key)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.New("key does not exist on server")
	}
	timestamp += 1
	// Ask storage to store
	// Note here the key is from get result
	if err := c.set(key, val, timestamp); err != nil {
		return "", err
	}
	return val, nil
}

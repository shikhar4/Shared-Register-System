package benchmark

import (
	"cs598fts/client"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const PAIR_CNT = 100000

type Workload int

var (
	workloadMap = map[string]Workload{
		"read-only":  ReadOnly,
		"write-only": WriteOnly,
		"half-half":  HalfHalf,
	}
)

func parseWorkloadStr(str string) (Workload, bool) {
	c, ok := workloadMap[strings.ToLower(str)]
	return c, ok
}

const (
	ReadOnly Workload = iota
	WriteOnly
	HalfHalf
)

type Benchmark struct {
	clientNum           int
	workload            Workload
	configPath          string
	requestCnt          int
	history             map[string]string
	lock                sync.Mutex
	validateCorrectness bool
}

func NewBenchmark(clientNum int, requestCnt int, workloadStr string, configPath string) (*Benchmark, error) {
	rand.Seed(time.Now().UnixNano())
	workload, ok := parseWorkloadStr(workloadStr)
	if !ok {
		return nil, errors.New("workload not supported")
	}

	benchmark := &Benchmark{
		clientNum:           clientNum,
		workload:            workload,
		configPath:          configPath,
		requestCnt:          requestCnt,
		history:             make(map[string]string),
		validateCorrectness: false,
	}

	return benchmark, nil
}

func (b *Benchmark) randomKey() string {
	return b.key(rand.Intn(PAIR_CNT))
}

func (b *Benchmark) randomVal() string {
	length := 10
	buf := make([]byte, length+2)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)[2 : length+2]
}

func (b *Benchmark) key(i int) string {
	return fmt.Sprintf("key-%d", i)
}

func (b *Benchmark) Init() error {
	logrus.Info("Init keys for benchmark")
	c, err := client.NewClient(0, b.configPath)
	if err != nil {
		return err
	}

	for i := 0; i < PAIR_CNT; i++ {
		key := b.key(i)
		val := b.randomVal()
		b.history[key] = val
		if err := c.Write(key, val); err != nil {
			return err
		}
	}

	return nil
}

func (b *Benchmark) read(c *client.Client) (string, error) {
	if b.validateCorrectness {
		b.lock.Lock()
		defer b.lock.Unlock()
	}
	key := b.randomKey()
	val, err := c.Read(key)
	if err != nil {
		return "", err
	}
	if b.validateCorrectness && val != b.history[key] {
		logrus.Debugf("%s, server: %s, local: %s", key, val, b.history[key])
	}
	return val, nil
}

func (b *Benchmark) write(c *client.Client) error {
	if b.validateCorrectness {
		b.lock.Lock()
		defer b.lock.Unlock()
	}
	key := b.randomKey()
	val := b.randomVal()
	if err := c.Write(key, val); err != nil {
		return err
	}
	if b.validateCorrectness {
		b.history[key] = val
	}
	return nil
}

func (b *Benchmark) doWorkload(c *client.Client) error {
	if b.workload == ReadOnly {
		if _, err := b.read(c); err != nil {
			return err
		}
		if _, err := b.read(c); err != nil {
			return err
		}
	} else if b.workload == WriteOnly {
		if err := b.write(c); err != nil {
			return err
		}
		if err := b.write(c); err != nil {
			return err
		}
	} else if b.workload == HalfHalf {
		if _, err := b.read(c); err != nil {
			return err
		}
		if err := b.write(c); err != nil {
			return err
		}
	}
	return nil
}

const PACKET_SIZE = 10

func opToByte(opCnt int) float64 {
	return float64(opCnt*PACKET_SIZE) / 1024.0
}

func (b *Benchmark) clientThread(clientID int, wg *sync.WaitGroup) {
	c, err := client.NewClient(clientID, b.configPath)
	if err != nil {
		logrus.Error(err)
		return
	}

	wantLog := clientID == 1

	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	totalOpCnt := 0
	totalSetCount := 0
	totalGetCount := 0
	var intervalOpCnt int32
	startTime := time.Now()

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				if wantLog {
					logrus.Debugf("[client %d] [%s] intervalOpCnt: %d, throughput: %f KB/s", clientID, t.String(), intervalOpCnt, opToByte(int(intervalOpCnt)))
				}
				atomic.StoreInt32(&intervalOpCnt, 0)
			}
		}
	}()

	for i := 0; i < (b.requestCnt+1)/2; i++ {
		if err := b.doWorkload(c); err != nil {
			logrus.Error(err)
		} // 2 ops
		totalOpCnt += 2
		totalSetCount += 2
		totalGetCount += 2
		atomic.AddInt32(&intervalOpCnt, 2)
	}

	done <- true
	runningTime := time.Now().Sub(startTime)
	logrus.Debugf("Finish benchmark workload on client %d, total time: %s, total operations: %d, avg throughput: %f KB/s, avg latency: %f ms/op", clientID, runningTime.String(), totalOpCnt, opToByte(totalOpCnt)/runningTime.Seconds(), float64(runningTime.Milliseconds())/float64(totalOpCnt))
	logrus.Infof("Finish benchmark workload on client %d, total sets done: %d, total gets done: %d, avg throughput: %f KB/s, avg latency: %f ms/op", clientID, totalSetCount, totalGetCount, opToByte(totalOpCnt)/runningTime.Seconds(), float64(runningTime.Milliseconds())/float64(totalOpCnt))
	wg.Done()
}

func (b *Benchmark) Run() error {
	logrus.Info("Running benchmark workload")
	var wg sync.WaitGroup
	wg.Add(b.clientNum)
	for i := 1; i <= b.clientNum; i++ {
		go b.clientThread(i, &wg)
	}
	wg.Wait()

	return nil
}

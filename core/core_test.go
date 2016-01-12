package core

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

const (
	retryCount    = 10
	diskAllocated = 1e9 // 1GB
)

func launchSwarm(nodeCount int, t *testing.T) []*server {
	port := 11100
	var nodes []*server

	var wg sync.WaitGroup
	wg.Add(nodeCount)

	for i := 0; i < nodeCount; i++ {
		port := port + i
		var peers []string
		if i > 0 {
			peers = append(peers, fmt.Sprintf("localhost:%d", nodes[i-1].network.Port))
		}
		s, err := newServer(port, peers, diskAllocated)
		if err != nil {
			t.Error(err)
		}
		go func() {
			wg.Done()
			if err := s.network.Listen(); err != nil {
				t.Error(err)
			}
		}()
		nodes = append(nodes, s)
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait()

	for i := 0; i < retryCount; i++ {
		var errors []error
		for _, node := range nodes {
			for _, peer := range nodes {
				protoPeer := peer.network.LocalPeer()
				if node.network.Peers[protoPeer.Id] == nil && node != peer {
					errors = append(errors, fmt.Errorf("node %+v missing peer %+v", node.network.LocalPeer(), protoPeer))
				}
			}
		}
		if len(errors) == 0 {
			break
		}
		if i < retryCount-1 {
			log.Printf("Rechecking peer discovery... %d times", retryCount-i)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, err := range errors {
			t.Error(err)
		}
	}
	return nodes
}

func TestCoreDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	launchSwarm(5, t)
}
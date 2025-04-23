package raftcore

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"

	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type RaftNode struct {
	Raft *raft.Raft
	FSM  *FSM
}

const ApplyTimeout = 5 * time.Second

// Config options passed into InitRaft
type RaftConfig struct {
	NodeID      string
	RaftDir     string
	BindAddr    string
	Peers       []string // Optional, can be empty
	IsBootstrap bool
}

// Initialize a Raft node
func InitRaft(config RaftConfig) (*RaftNode, error) {
	// Create the Raft config
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.NodeID)

	// Create Raft data directory if not present
	if err := os.MkdirAll(config.RaftDir, 0700); err != nil {
		return nil, err
	}

	// Setup Raft communication transport
	addr, err := net.ResolveTCPAddr("tcp", config.BindAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(config.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the snapshot store
	snapshots, err := raft.NewFileSnapshotStore(config.RaftDir, 1, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the BoltDB log store
	logStorePath := filepath.Join(config.RaftDir, "raft-log.bolt")
	logStore, err := raftboltdb.NewBoltStore(logStorePath)
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}

	// Create the FSM
	fsm := NewFSM()

	// Initialize the Raft system
	r, err := raft.NewRaft(raftConfig, fsm, logStore, logStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
	}

	// Bootstrap the cluster only on the first node
	if config.IsBootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(config.NodeID),
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(config)
	}

	return &RaftNode{
		Raft: r,
		FSM:  fsm,
	}, nil
}

// Utility function to list peers (can be used in API)
func (rn *RaftNode) ListPeers() []raft.Server {
	return rn.Raft.GetConfiguration().Configuration().Servers
}

// IsLeader returns true if this node is the current leader
func (rn *RaftNode) IsLeader() bool {
	return rn.Raft.State() == raft.Leader
}

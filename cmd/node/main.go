package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"raft3d/api"
	"raft3d/raftcore"
)

const ApplyTimeout = 5 * time.Second

func main() {
	var nodeID string
	var httpAddr string
	var raftAddr string
	var raftDir string
	var bootstrap bool

	flag.StringVar(&nodeID, "id", "node1", "Node ID")
	flag.StringVar(&httpAddr, "http", "localhost:9001", "HTTP bind address")
	flag.StringVar(&raftAddr, "raft", "localhost:12001", "Raft bind address")
	flag.StringVar(&raftDir, "data", "", "Data directory for Raft")
	flag.BoolVar(&bootstrap, "bootstrap", false, "Bootstrap the cluster")

	flag.Parse()

	if raftDir == "" {
		raftDir = fmt.Sprintf("raft-data-%s", nodeID)
	}

	os.MkdirAll(raftDir, 0700)

	// Init Raft
	raftNode, err := raftcore.InitRaft(raftcore.RaftConfig{
		NodeID:      nodeID,
		RaftDir:     raftDir,
		BindAddr:    raftAddr,
		IsBootstrap: bootstrap,
	})
	if err != nil {
		log.Fatalf("Failed to start Raft: %v", err)
	}

	// Start API server
	server := api.NewServer(raftNode)

	http.HandleFunc("/api/v1/printers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			server.ListPrinters(w, r)
		} else if r.Method == "POST" {
			server.CreatePrinter(w, r)
		}
	})

	http.HandleFunc("/api/v1/filaments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			server.ListFilaments(w, r)
		} else if r.Method == "POST" {
			server.CreateFilament(w, r)
		}
	})

	http.HandleFunc("/api/v1/print_jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			server.ListPrintJobs(w, r)
		} else if r.Method == "POST" {
			server.CreatePrintJob(w, r)
		}
	})

	http.HandleFunc("/api/v1/print_jobs/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			server.UpdatePrintJobStatus(w, r)
		}
	})
	http.HandleFunc("/api/v1/join", server.JoinNode)

	log.Printf("[%s] HTTP API listening on %s", nodeID, httpAddr)
	log.Printf("[%s] Raft state: %s", nodeID, raftNode.Raft.State())
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

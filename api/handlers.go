package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"raft3d/models"
	"raft3d/raftcore"

	"github.com/hashicorp/raft"
)

type Server struct {
	RaftNode *raftcore.RaftNode
}

func NewServer(rn *raftcore.RaftNode) *Server {
	return &Server{RaftNode: rn}
}

// Helper to parse JSON
func parseBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

// POST /api/v1/printers
func (s *Server) CreatePrinter(w http.ResponseWriter, r *http.Request) {
	var printer models.Printer
	if !parseBody(w, r, &printer) {
		return
	}

	fmt.Println("Received printer create request:", printer)

	if !s.RaftNode.IsLeader() {
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	printerData, _ := json.Marshal(printer)
	cmd := raftcore.FSMCommand{
		Type: "printer",
		Data: printerData,
	}

	data, _ := json.Marshal(cmd)
	future := s.RaftNode.Raft.Apply(data, raftcore.ApplyTimeout)
	if err := future.Error(); err != nil {
		http.Error(w, "Apply failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(printer)
}

// GET /api/v1/printers
func (s *Server) ListPrinters(w http.ResponseWriter, r *http.Request) {
	s.RaftNode.FSM.Lock()
	defer s.RaftNode.FSM.Unlock()

	list := make([]models.Printer, 0, len(s.RaftNode.FSM.Printers))
	for _, p := range s.RaftNode.FSM.Printers {
		list = append(list, p)
	}

	json.NewEncoder(w).Encode(list)
}

// POST /api/v1/filaments
func (s *Server) CreateFilament(w http.ResponseWriter, r *http.Request) {
	var filament models.Filament
	if !parseBody(w, r, &filament) {
		return
	}

	fmt.Println("Received filament create request:", filament)

	if !s.RaftNode.IsLeader() {
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	filamentData, _ := json.Marshal(filament)
	cmd := raftcore.FSMCommand{
		Type: "filament",
		Data: filamentData,
	}

	fmt.Println("Applying filament to Raft")

	data, _ := json.Marshal(cmd)
	future := s.RaftNode.Raft.Apply(data, raftcore.ApplyTimeout)
	if err := future.Error(); err != nil {
		http.Error(w, "Apply failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(filament)
}

// GET /api/v1/filaments
func (s *Server) ListFilaments(w http.ResponseWriter, r *http.Request) {
	s.RaftNode.FSM.Lock()
	defer s.RaftNode.FSM.Unlock()

	list := make([]models.Filament, 0, len(s.RaftNode.FSM.Filaments))
	for _, f := range s.RaftNode.FSM.Filaments {
		list = append(list, f)
	}

	json.NewEncoder(w).Encode(list)
}

// POST /api/v1/print_jobs
func (s *Server) CreatePrintJob(w http.ResponseWriter, r *http.Request) {
	var job models.PrintJob
	if !parseBody(w, r, &job) {
		return
	}

	if !s.RaftNode.IsLeader() {
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	// Basic validation (printer and filament must exist)
	s.RaftNode.FSM.Lock()
	defer s.RaftNode.FSM.Unlock()

	_, ok1 := s.RaftNode.FSM.Printers[job.PrinterID]
	_, ok2 := s.RaftNode.FSM.Filaments[job.FilamentID]
	if !ok1 || !ok2 {
		http.Error(w, "Invalid printer or filament ID", http.StatusBadRequest)
		return
	}

	jobData, _ := json.Marshal(job)
	cmd := raftcore.FSMCommand{
		Type: "print_job",
		Data: jobData,
	}
	data, _ := json.Marshal(cmd)
	fmt.Println("Submitting print job to Raft:", job)

	future := s.RaftNode.Raft.Apply(data, raftcore.ApplyTimeout)
	if err := future.Error(); err != nil {
		http.Error(w, "Apply failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GET /api/v1/print_jobs
func (s *Server) ListPrintJobs(w http.ResponseWriter, r *http.Request) {
	s.RaftNode.FSM.Lock()
	defer s.RaftNode.FSM.Unlock()

	list := make([]models.PrintJob, 0, len(s.RaftNode.FSM.PrintJobs))
	for _, j := range s.RaftNode.FSM.PrintJobs {
		list = append(list, j)
	}

	json.NewEncoder(w).Encode(list)
}

// POST /api/v1/print_jobs/{id}/status?status=done
func (s *Server) UpdatePrintJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	status := r.URL.Query().Get("status")
	if id == "" || status == "" {
		http.Error(w, "Missing id or status", http.StatusBadRequest)
		return
	}

	if !s.RaftNode.IsLeader() {
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"id":     id,
		"status": status,
	})
	cmd := raftcore.FSMCommand{
		Type: "update_status",
		Data: updateData,
	}
	data, _ := json.Marshal(cmd)
	future := s.RaftNode.Raft.Apply(data, raftcore.ApplyTimeout)
	if err := future.Error(); err != nil {
		http.Error(w, "Apply failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Status updated")
}

// POST /api/v1/join?id=node2&addr=127.0.0.1:12002
func (s *Server) JoinNode(w http.ResponseWriter, r *http.Request) {
	if !s.RaftNode.IsLeader() {
		http.Error(w, "Only leader can accept joins", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	addr := r.URL.Query().Get("addr")
	if id == "" || addr == "" {
		http.Error(w, "Missing id or addr", http.StatusBadRequest)
		return
	}

	future := s.RaftNode.Raft.AddVoter(
		raft.ServerID(id),
		raft.ServerAddress(addr),
		0, // prevIndex (0 = any)
		raftcore.ApplyTimeout,
	)
	if err := future.Error(); err != nil {
		http.Error(w, "Failed to add voter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Node %s at %s joined successfully", id, addr)
}

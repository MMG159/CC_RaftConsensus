package raftcore

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"sync"

	"raft3d/models"

	"github.com/hashicorp/raft"
)

// FSM implements the raft.FSM interface
type FSM struct {
	mu        sync.Mutex
	Printers  map[string]models.Printer
	Filaments map[string]models.Filament
	PrintJobs map[string]models.PrintJob
}

// FSMCommand is used to wrap Raft log commands
type FSMCommand struct {
	Type string          `json:"type"` // printer, filament, job, update_status
	Data json.RawMessage `json:"data"`
}

// NewFSM creates a new FSM instance
func NewFSM() *FSM {
	return &FSM{
		Printers:  make(map[string]models.Printer),
		Filaments: make(map[string]models.Filament),
		PrintJobs: make(map[string]models.PrintJob),
	}
}

// Apply applies a Raft log entry to the FSM.
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	var cmd FSMCommand
	if err := json.Unmarshal(logEntry.Data, &cmd); err != nil {
		log.Printf("Failed to unmarshal log: %v", err)
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case "printer":
		var p models.Printer
		if err := json.Unmarshal(cmd.Data, &p); err == nil {
			f.Printers[p.ID] = p
			log.Printf("Added printer: %+v", p)
		}
	case "filament":
		var fil models.Filament
		if err := json.Unmarshal(cmd.Data, &fil); err == nil {
			f.Filaments[fil.ID] = fil
			log.Printf("Added filament: %+v", fil)
		}
	case "print_job":
		var job models.PrintJob
		if err := json.Unmarshal(cmd.Data, &job); err == nil {
			job.Status = "Queued"
			f.PrintJobs[job.ID] = job
			log.Printf("Added print job: %+v", job)
		}
	case "update_status":
		var update struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(cmd.Data, &update); err == nil {
			job := f.PrintJobs[update.ID]

			if job.Status == "Queued" && update.Status == "Running" ||
				job.Status == "Running" && update.Status == "Done" ||
				(job.Status == "Queued" || job.Status == "Running") && update.Status == "Cancelled" {

				job.Status = update.Status

				if update.Status == "Done" {
					fil := f.Filaments[job.FilamentID]
					fil.RemainingWeightInGrams -= job.PrintWeightInGrams
					f.Filaments[job.FilamentID] = fil
				}

				f.PrintJobs[update.ID] = job
				log.Printf("Updated job status: %s -> %s", update.ID, update.Status)
			}
		}
	default:
		log.Printf("Unknown command type: %s", cmd.Type)
	}

	return nil
}

// Snapshot returns a snapshot of the FSM.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return &fsmSnapshot{state: data}, nil
}

// Restore restores the FSM from a snapshot.
func (f *FSM) Restore(rc io.ReadCloser) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return err
	}

	var snapshot FSM
	if err := json.Unmarshal(buf.Bytes(), &snapshot); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.Printers = snapshot.Printers
	f.Filaments = snapshot.Filaments
	f.PrintJobs = snapshot.PrintJobs

	return nil
}

// fsmSnapshot implements raft.FSMSnapshot
type fsmSnapshot struct {
	state []byte
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(s.state); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}

// Lock/Unlock for external access
func (f *FSM) Lock() {
	f.mu.Lock()
}

func (f *FSM) Unlock() {
	f.mu.Unlock()
}

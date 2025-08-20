package process

// ProcStatus is a wrapper with the process current status.
type ProcStatus struct {
	Status   string // "running", "stopped", "exited"
	Restarts int    // number of restarts
}

// SetStatus will set the process string status.
func (p *ProcStatus) SetStatus(status string) {
	p.Status = status
}

// AddRestart will add one restart to the process status.
func (p *ProcStatus) AddRestart() {
	p.Restarts++
}

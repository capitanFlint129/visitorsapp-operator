package workload_ensurers

type ensureWorkloadDirector struct {
	ensurer WorkloadEnsurer
}

func (e *ensureWorkloadDirector) SetEnsurer(ensurer WorkloadEnsurer) {
	e.ensurer = ensurer
}

func (e *ensureWorkloadDirector) EnsureMySQL() {

}

func (e *ensureWorkloadDirector) EnsureBackend() {

}

func (e *ensureWorkloadDirector) EnsureFrontend() {

}

func NewEnsureWorkloadDirector() EnsureWorkloadDirector {
	return &ensureWorkloadDirector{}
}

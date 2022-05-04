package workload_ensurers

type EnsureWorkloadDirector = interface {
	SetEnsurer(ensurer WorkloadEnsurer)
	EnsureMySQL()
	EnsureBackend()
	EnsureFrontend()
}

type WorkloadEnsurer = interface {
	EnsureDeployment()
	EnsureService()
	EnsureSecret()
}

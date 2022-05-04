package workload_ensurers

type backendEnsurer struct {
}

func (b *backendEnsurer) EnsureDeployment() {

}

func (b *backendEnsurer) EnsureService() {

}

func (b *backendEnsurer) EnsureSecret() {

}

func NewBackendEnsurer() WorkloadEnsurer {
	return &backendEnsurer{}
}

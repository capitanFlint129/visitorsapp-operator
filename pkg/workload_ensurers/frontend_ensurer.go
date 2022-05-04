package workload_ensurers

type frontendEnsurer struct {
}

func (b *frontendEnsurer) EnsureDeployment() {

}

func (b *frontendEnsurer) EnsureService() {

}

func (b *frontendEnsurer) EnsureSecret() {

}

func NewFrontendEnsurer() WorkloadEnsurer {
	return &frontendEnsurer{}
}

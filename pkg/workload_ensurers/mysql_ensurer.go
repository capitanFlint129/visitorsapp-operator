package workload_ensurers

type mysqlEnsurer struct {
}

func (b *mysqlEnsurer) EnsureDeployment() {

}

func (b *mysqlEnsurer) EnsureService() {

}

func (b *mysqlEnsurer) EnsureSecret() {

}

func NewMysqlEnsurer() WorkloadEnsurer {
	return &mysqlEnsurer{}
}

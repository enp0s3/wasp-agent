package args

// FactoryArgs contains the required parameters to generate all namespaced resources
type FactoryArgs struct {
	OperatorVersion        string `required:"true" split_words:"true"`
	WaspImage              string `required:"true" split_words:"true"`
	DeployClusterResources string `required:"true" split_words:"true"`
	DeployPrometheusRule   string `required:"true" split_words:"true"`
	Verbosity              string `required:"true"`
	PullPolicy             string `required:"true" split_words:"true"`
	Namespace              string
}

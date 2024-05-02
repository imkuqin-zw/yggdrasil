package trace

type Config struct {
	// ExcludeQueryVars configures the db.statement attribute to exclude query variables
	ExcludeQueryVars bool `yaml:"excludeQueryVars" json:"excludeQueryVars"`
}

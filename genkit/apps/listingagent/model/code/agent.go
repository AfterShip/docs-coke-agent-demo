package code

// Agent errors. module code = 01
const (
	// ErrSpannerUnknown - 500: Spanner unknown error.
	ErrSpannerUnknown uint32 = iota + SystemCode + AgentModelCode*1000

	// ErrESUnknown - 500: ElasticSearch unknown error.
	ErrESUnknown
)

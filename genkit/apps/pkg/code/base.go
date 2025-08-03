package code

// SystemCode 系统编码，「服务」是针对这整个系统而言，比如 apiserver + worker 是一个系统，虽然是两个服务。
const SystemCode uint32 = 100 * 100 * 1001

const (
	// Success - 200: OK.
	Success uint32 = 0
)

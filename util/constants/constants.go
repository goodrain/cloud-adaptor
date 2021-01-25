package constants

const (
	// Namespace used by the micro API
	Namespace = "com.goodrain"
	// Service is the suffix of FullService
	Service = "enterprise-server"
	// FullService is service name used to micro registry
	FullService = Namespace + "." + Service
	// CloudInit Cloud resource initialization constant
	CloudInit = "cloud-init"
	// CloudCreate Cloud resource create constant
	CloudCreate = "cloud-create"
)

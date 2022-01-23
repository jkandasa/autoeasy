package types

const (
	// functions
	FuncAdd          = "add"
	FuncRemove       = "remove"
	FuncRemoveAll    = "remove_all"
	FuncKeepOnly     = "keep_only"
	FuncPrintInfo    = "print_info"
	FuncWaitForReady = "wait_for_ready"

	// kinds
	KindSubscription             = "Subscription"
	KindImageContentSourcePolicy = "ImageContentSourcePolicy"
	KindCatalogSource            = "CatalogSource"
	KindNamespace                = "Namespace"
	KindDeployment               = "Deployment"
	KindRoute                    = "Route"
)

package types

const (
	// functions
	FuncGet           = "get"
	FuncAdd           = "add"
	FuncRemove        = "remove"
	FuncRemoveAll     = "remove_all"
	FuncKeepOnly      = "keep_only"
	FuncPrintInfo     = "print_info"
	FuncWaitForReady  = "wait_for_ready"
	FuncWaitForDelete = "wait_for_delete"
	FuncLogin         = "login"
	FuncLogout        = "logout"

	// kinds
	KindSubscription             = "Subscription"
	KindImageContentSourcePolicy = "ImageContentSourcePolicy"
	KindCatalogSource            = "CatalogSource"
	KindNamespace                = "Namespace"
	KindDeployment               = "Deployment"
	KindPod                      = "Pod"
	KindRoute                    = "Route"
	KindInternal                 = "Internal"
)

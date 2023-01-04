package init

import (
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/catalog_source"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/create"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/delete"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/get"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/icsp"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/install"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/operator"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/registry"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/uninstall"
)

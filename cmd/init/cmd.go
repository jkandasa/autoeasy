package init

import (
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/catalog_source"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/get"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/install"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/operator"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/registry"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"
	_ "github.com/jkandasa/autoeasy/cmd/plugin/openshift/uninstall"
)

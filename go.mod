module software-update-manager

go 1.12

replace plugin-manager => ../plugin-manager/

require (
	gopkg.in/yaml.v2 v2.4.0
	plugin-manager v1.4.6
)

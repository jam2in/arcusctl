package config

const (
	ArcusRootPath               = "/arcus"
	AclRootPath                 = "/arcus_acl"
	ArcusCacheListPath          = "/arcus/cache_list"
	ArcusClientListPath         = "/arcus/client_list"
	ArcusCacheServerMappingPath = "/arcus/cache_server_mapping"

	PropName = "authPassword"
)

var ArcusBasicPaths = []string{
	ArcusRootPath,
	AclRootPath,
	ArcusCacheListPath,
	ArcusClientListPath,
	ArcusCacheServerMappingPath,
}

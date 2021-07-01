package config

type Libp2pConf struct {
	BootstrapNodes []string
	ListenAddrs    []string
	MDnsName       string
}

type DatabaseConfig struct {
	Path string
}

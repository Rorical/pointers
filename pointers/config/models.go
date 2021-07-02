package config

type Libp2pConf struct {
	BootstrapNodes []string
	ListenAddrs    []string
	GroupName      string
}

type DatabaseConfig struct {
	Path string
}

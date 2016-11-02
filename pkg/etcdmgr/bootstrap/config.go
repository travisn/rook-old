package bootstrap

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"time"

	"github.com/coreos/etcd/pkg/types"
)

const (
	// DefaultClientPort is the default port for listening to client traffic
	DefaultClientPort = 53379
	// DefaultPeerPort is the default port for listening to peer traffic
	DefaultPeerPort = 53380
	// DefaultClientTimeout is the default timeout for etcd client
	DefaultClientTimeout = 5 * time.Second
)

// Config holds the configuration of elastic etcd.
type Config struct {
	InstanceName        string
	ListenPeerURLs      []url.URL
	ListenClientURLs    []url.URL
	AdvertisePeerURLs   []url.URL
	AdvertiseClientURLs []url.URL
	DataDir             string
	URLsMap             types.URLsMap
}

// CopyConfig returns a copy of given etcd config
func CopyConfig(conf *Config) Config {
	newConf := Config{
		InstanceName:        conf.InstanceName,
		ListenPeerURLs:      conf.ListenPeerURLs,
		ListenClientURLs:    conf.ListenClientURLs,
		AdvertisePeerURLs:   conf.AdvertisePeerURLs,
		AdvertiseClientURLs: conf.AdvertiseClientURLs,
		DataDir:             conf.DataDir,
		URLsMap:             conf.URLsMap,
	}

	return newConf
}

// newConfig automatically generates a config for embedded etcd
func newConfig(configDir, ipAddr, nodeID string) *Config {
	conf := &Config{}
	conf.InstanceName = nodeID
	conf.ListenPeerURLs = append(conf.ListenPeerURLs,
		getURLFromSchemeIPPort("http", ipAddr, DefaultPeerPort))
	conf.ListenClientURLs = append(conf.ListenClientURLs,
		getURLFromSchemeIPPort("http", ipAddr, DefaultClientPort))
	conf.AdvertisePeerURLs = append(conf.AdvertisePeerURLs,
		getURLFromSchemeIPPort("http", ipAddr, DefaultPeerPort))
	conf.AdvertiseClientURLs = append(conf.AdvertiseClientURLs,
		getURLFromSchemeIPPort("http", ipAddr, DefaultClientPort))
	conf.DataDir = path.Join(configDir, "etcd-data")
	conf.URLsMap = types.URLsMap{}

	log.Printf("etcd config: %+v", conf)

	return conf
}

// GenerateConfigFromExistingCluster automatically generates a config for joining an existing cluster
func GenerateConfigFromExistingCluster(Context EtcdMgrContext, configDir, ipAddr, nodeID string) (*Config, error) {
	conf := newConfig(configDir, ipAddr, nodeID)

	// get current urlmap of the etcd cluster
	var err error
	_, conf.URLsMap, err = Context.Members()
	if err != nil {
		return nil, err
	}

	// add the entry for the new member
	conf.URLsMap[conf.InstanceName] = conf.AdvertisePeerURLs
	log.Println("conf.URLsMap: ", conf.URLsMap)
	return conf, nil
}

func getURLFromSchemeIPPort(scheme, ip string, port int) url.URL {
	u := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", ip, port),
	}
	return u
}

package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
	"gopkg.in/yaml.v3"

	"github.com/BurntSushi/toml"
	"github.com/enjoy-vc/router/common/xnet"
	_ "github.com/xjasonlyu/tun2socks/v2/dns"
	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xjasonlyu/tun2socks/v2/internal/version"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

var (
	key = new(engine.Key)

	configFile  string
	configUrl   string
	versionFlag bool
)

func init() {
	// flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux only)")
	// flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.DurationVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.StringVar(&configUrl, "config-url", "", "xmit config url")
	flag.StringVar(&configFile, "config", "", "YAML/TOML format configuration file")
	flag.StringVar(&key.Device, "device", "", "Use this device [driver://]name")
	// flag.StringVar(&key.Interface, "interface", "", "Use network INTERFACE (Linux/MacOS only)")
	// flag.StringVar(&key.LogLevel, "loglevel", "info", "Log level [debug|info|warning|error|silent]")
	// flag.StringVar(&key.Proxy, "proxy", "", "Use this proxy [protocol://]host[:port]")
	// flag.StringVar(&key.RestAPI, "restapi", "", "HTTP statistic server listen address")
	// flag.StringVar(&key.TCPSendBufferSize, "tcp-sndbuf", "", "Set TCP send buffer size for netstack")
	// flag.StringVar(&key.TCPReceiveBufferSize, "tcp-rcvbuf", "", "Set TCP receive buffer size for netstack")
	// flag.BoolVar(&key.TCPModerateReceiveBuffer, "tcp-auto-tuning", false, "Enable TCP receive buffer auto-tuning")
	// flag.StringVar(&key.MulticastGroups, "multicast-groups", "", "Set multicast groups, separated by commas")
	// flag.StringVar(&key.TUNPreUp, "tun-pre-up", "", "Execute a command before TUN device setup")
	// flag.StringVar(&key.TUNPostUp, "tun-post-up", "", "Execute a command after TUN device setup")
	flag.BoolVar(&versionFlag, "version", false, "Show version and then quit")
	flag.Parse()
}

func main() {
	maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))

	if versionFlag {
		fmt.Println(version.String())
		fmt.Println(version.BuildString())
		os.Exit(0)
	}

	if configUrl != "" {
		if err := queryConfig(configUrl, key); err != nil {
			log.Fatalf("Failed to read config url '%s': %v", configUrl, err)
		}
	} else if configFile != "" {
		if false {
			// 通过web拿数据
		} else {
			data, err := os.ReadFile(configFile)
			if err != nil {
				log.Fatalf("Failed to read config file '%s': %v", configFile, err)
			}
			ext := path.Ext(configFile)
			switch ext {
			case ".yaml":
				if err = yaml.Unmarshal(data, key); err != nil {
					log.Fatalf("Failed to unmarshal config file '%s': %v", configFile, err)
				}
			case ".toml":
				if err = decodeToml(data, key); err != nil {
					log.Fatalf("Failed to unmarshal config file '%s': %v", configFile, err)
				}
			default:
				log.Fatalf("unknown ext: %v", ext)
			}

		}
	}

	engine.Insert(key)

	engine.Start()
	defer engine.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func decodeToml(data []byte, key *engine.Key) error {
	var tmp struct {
		Tun engine.Key
	}
	tmp.Tun = engine.Key{
		Proxy: "xmit://zfliu:940625@127.0.0.1:7890",
	}
	_, err := toml.Decode(string(data), &tmp)
	if err != nil {
		return err
	}
	proxy.SetXmitDefaultConfig(data)
	*key = tmp.Tun
	return nil
}

func queryConfig(baseUrl string, key *engine.Key) error {
	itf, err := net.InterfaceByName("eth0")
	if err != nil {
		if err.Error() == "route ip+net: no such network interface" {
			itf, err = net.InterfaceByName("ens160")
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	result, err := url.JoinPath(baseUrl, "servercpe/api/cpe/update_cpe_xmit")
	if err != nil {
		return err
	}
	result = fmt.Sprintf("%v?macaddr=%v-xmit", result, itf.HardwareAddr.String())
	respon, err := xnet.DoHttpGet(result)
	if err != nil {
		return err
	}
	fmt.Println(string(respon))

	return nil
}

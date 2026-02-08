package mikrotik_outbound_adapter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-routeros/routeros/v3"

	"mikrops/internal/model"
)

type RouterOSClient struct {
	address  string
	username string
	password string
}

func NewRouterOSClient(host string, port int, username, password string) *RouterOSClient {
	return &RouterOSClient{
		address:  fmt.Sprintf("%s:%d", host, port),
		username: username,
		password: password,
	}
}

func (c *RouterOSClient) connect() (*routeros.Client, error) {
	client, err := routeros.DialTimeout(c.address, c.username, c.password, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *RouterOSClient) TestConnection() error {
	client, err := c.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

func (c *RouterOSClient) GetIdentity() (string, error) {
	client, err := c.connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	reply, err := client.Run("/system/identity/print")
	if err != nil {
		return "", err
	}

	if len(reply.Re) > 0 {
		if name, ok := reply.Re[0].Map["name"]; ok {
			return name, nil
		}
	}

	return "", fmt.Errorf("identity not found")
}

func (c *RouterOSClient) GetActiveConnections() ([]model.MikrotikConnection, error) {
	client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ppp/active/print")
	if err != nil {
		return nil, err
	}

	connections := make([]model.MikrotikConnection, 0, len(reply.Re))
	for _, re := range reply.Re {
		connection := model.MikrotikConnection{
			ID:       re.Map[".id"],
			Name:     re.Map["name"],
			Service:  re.Map["service"],
			CallerID: re.Map["caller-id"],
			Address:  re.Map["address"],
			Uptime:   re.Map["uptime"],
		}
		connections = append(connections, connection)
	}

	return connections, nil
}

func (c *RouterOSClient) GetInterfaceTraffic(interfaceName string) (model.MikrotikTraffic, error) {
	client, err := c.connect()
	if err != nil {
		return model.MikrotikTraffic{}, err
	}
	defer client.Close()

	reply, err := client.Run("/interface/print", "?name="+interfaceName)
	if err != nil {
		return model.MikrotikTraffic{}, err
	}

	if len(reply.Re) == 0 {
		return model.MikrotikTraffic{}, fmt.Errorf("interface %s not found", interfaceName)
	}

	re := reply.Re[0].Map
	traffic := model.MikrotikTraffic{
		Interface: interfaceName,
		RxBytes:   parseInt64(re["rx-byte"]),
		TxBytes:   parseInt64(re["tx-byte"]),
		RxPackets: parseInt64(re["rx-packet"]),
		TxPackets: parseInt64(re["tx-packet"]),
	}

	return traffic, nil
}

func (c *RouterOSClient) MonitorBandwidth(target string) (model.MikrotikBandwidth, error) {
	client, err := c.connect()
	if err != nil {
		return model.MikrotikBandwidth{}, err
	}
	defer client.Close()

	// Monitor queue for the target
	reply, err := client.Run("/queue/simple/print", "?target="+target)
	if err != nil {
		return model.MikrotikBandwidth{}, err
	}

	if len(reply.Re) == 0 {
		return model.MikrotikBandwidth{}, fmt.Errorf("queue for target %s not found", target)
	}

	re := reply.Re[0].Map
	bandwidth := model.MikrotikBandwidth{
		Target:   target,
		Upload:   parseInt64(re["rate"]),
		Download: parseInt64(re["rate"]),
	}

	return bandwidth, nil
}

// Helper function to parse int64
func parseInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

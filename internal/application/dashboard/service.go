package dashboard

import (
	proxyapp "podux/internal/application/proxy"
	serverapp "podux/internal/application/server"
	proxydomain "podux/internal/domain/proxy"
	tunneldomain "podux/internal/domain/tunnel"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type DashboardStats struct {
	RunningCount         int64
	StoppedCount         int64
	ProxiesEnabledCount  int64
	ProxiesDisabledCount int64
	ProxiesOnlineCount   int64
	MaxLatency           int64
	ProxyTypeCounts      map[string]int64
	UptimeSeconds        int64
}

type TopologyServerItem struct {
	ID           string `json:"id" db:"id"`
	ServerName   string `json:"serverName" db:"serverName"`
	ServerAddr   string `json:"serverAddr" db:"serverAddr"`
	TunnelStatus string `json:"bootStatus" db:"tunnelStatus"`
}

type TopologyProxyItem struct {
	ID         string `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	ProxyType  string `json:"proxyType" db:"proxyType"`
	LocalIP    string `json:"localIP" db:"localIP"`
	LocalPort  string `json:"localPort" db:"localPort"`
	RemotePort string `json:"remotePort" db:"remotePort"`
	Status     string `json:"status" db:"status"`
	BootStatus string `json:"bootStatus" db:"bootStatus"`
	ServerID   string `json:"serverId" db:"serverId"`
}

type TopologyData struct {
	Servers []TopologyServerItem `json:"servers"`
	Proxies []TopologyProxyItem  `json:"proxies"`
}

type Service struct {
	app           core.App
	serverService *serverapp.Service
	proxyService  *proxyapp.Service
	tunnelRepo    tunneldomain.Repository
	startTime     time.Time
}

func NewService(app core.App, serverService *serverapp.Service, proxyService *proxyapp.Service, tunnelRepo tunneldomain.Repository) *Service {
	return &Service{
		app:           app,
		serverService: serverService,
		proxyService:  proxyService,
		tunnelRepo:    tunnelRepo,
		startTime:     time.Now(),
	}
}

func (s *Service) GetStats() (*DashboardStats, error) {
	// Running/stopped server counts derived from FRP tunnel statuses.
	var runningCount, stoppedCount int64
	s.app.DB().Select("COUNT(*)").From("th_tunnels").
		Where(dbx.And(dbx.HashExp{"type": "frp"}, dbx.HashExp{"status": string(tunneldomain.TunnelStatusActive)})).
		Row(&runningCount)
	s.app.DB().Select("COUNT(*)").From("th_tunnels").
		Where(dbx.And(dbx.HashExp{"type": "frp"}, dbx.HashExp{"status": string(tunneldomain.TunnelStatusInactive)})).
		Row(&stoppedCount)

	proxiesEnabledCount, err := s.proxyService.GetProxyCountByStatus(proxydomain.ProxyStatusEnabled)
	if err != nil {
		return nil, err
	}

	proxiesDisabledCount, err := s.proxyService.GetProxyCountByStatus(proxydomain.ProxyStatusDisabled)
	if err != nil {
		return nil, err
	}

	proxiesOnlineCount, err := s.proxyService.GetProxyCountByBootStatus(proxydomain.ProxyBootStatusOnline)
	if err != nil {
		return nil, err
	}

	maxLatency, err := s.serverService.GetMaxLatency()
	if err != nil {
		return nil, err
	}

	proxyTypeCounts, err := s.proxyService.GetProxyCountsByType()
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		RunningCount:         runningCount,
		StoppedCount:         stoppedCount,
		ProxiesEnabledCount:  proxiesEnabledCount,
		ProxiesDisabledCount: proxiesDisabledCount,
		ProxiesOnlineCount:   proxiesOnlineCount,
		MaxLatency:           maxLatency,
		ProxyTypeCounts:      proxyTypeCounts,
		UptimeSeconds:        int64(time.Since(s.startTime).Seconds()),
	}, nil
}

func (s *Service) GetTopology() (*TopologyData, error) {
	// Servers: join with th_tunnels (type=frp) to get running status.
	var servers []TopologyServerItem
	err := s.app.DB().
		Select("fs.id", "fs.serverName", "fs.serverAddr",
			"COALESCE(tt.status, 'inactive') as tunnelStatus").
		From("fh_servers fs").
		LeftJoin("th_tunnels tt", dbx.NewExp("tt.serverId = fs.id AND tt.serverId IS NOT NULL AND tt.serverId != ''")).
		All(&servers)
	if err != nil {
		return nil, err
	}
	if servers == nil {
		servers = []TopologyServerItem{}
	}

	// Proxies: bootStatus still comes from fh_proxies directly.
	var proxies []TopologyProxyItem
	if err := s.app.DB().
		Select("id", "name", "proxyType", "localIP", "localPort", "remotePort", "status", "bootStatus", "serverId").
		From("fh_proxies").
		All(&proxies); err != nil {
		return nil, err
	}
	if proxies == nil {
		proxies = []TopologyProxyItem{}
	}

	return &TopologyData{Servers: servers, Proxies: proxies}, nil
}

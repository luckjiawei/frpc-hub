package httphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"podux/internal/application/scanner"
	"podux/pkg/response"

	"github.com/pocketbase/pocketbase/core"
)

type ScannerHandler struct {
	app     core.App
	service *scanner.Service
}

func NewScannerHandler(app core.App, service *scanner.Service) *ScannerHandler {
	return &ScannerHandler{app: app, service: service}
}

func (h *ScannerHandler) RegisterHandlers(e *core.ServeEvent) {
	// GET /api/scanner/interfaces
	// Returns local networks that will be scanned.
	e.Router.GET("/api/scanner/interfaces", requireAuth(func(e *core.RequestEvent) error {
		return e.JSON(200, h.service.GetLocalNetworks())
	}, h.app))

	// GET /api/scanner/discover
	// ARP-sweep local network, returns discovered hosts.
	e.Router.GET("/api/scanner/discover", requireAuth(func(e *core.RequestEvent) error {
		ctx := e.Request.Context()
		devices, err := h.service.DiscoverDevices(ctx)
		if err != nil {
			h.app.Logger().Error("LAN discovery failed", "error", err)
			return e.JSON(500, response.Error(err))
		}
		return e.JSON(200, devices)
	}, h.app))

	// GET /api/scanner/ports/stream?ip=<ip>[&ports=<p1,p2,...>]
	// Streams port-scan results as Server-Sent Events.
	e.Router.GET("/api/scanner/ports/stream", requireAuth(func(e *core.RequestEvent) error {
		q := e.Request.URL.Query()
		ip := q.Get("ip")
		if ip == "" {
			return e.JSON(400, response.Error(fmt.Errorf("ip is required")))
		}
		if net.ParseIP(ip) == nil {
			return e.JSON(400, response.Error(fmt.Errorf("invalid ip address")))
		}

		var ports []int
		switch {
		case q.Get("full") == "true":
			ports = allPorts()
		case q.Get("ports") != "":
			ports = parsePortList(q.Get("ports"))
		default:
			ports = scanner.CommonPorts
		}

		w := e.Response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			return e.JSON(500, response.Error(fmt.Errorf("streaming not supported")))
		}

		ctx, cancel := context.WithCancel(e.Request.Context())
		defer cancel()

		results := make(chan scanner.PortResult, 128)

		go func() {
			defer close(results)
			h.service.ScanPorts(ctx, ip, ports, results)
		}()

		for res := range results {
			data, err := json.Marshal(res)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		// Signal completion
		done, _ := json.Marshal(scanner.PortResult{Done: true})
		fmt.Fprintf(w, "data: %s\n\n", done)
		flusher.Flush()

		return nil
	}, h.app))
}

// allPorts returns every valid TCP port (1-65535).
func allPorts() []int {
	ports := make([]int, 65535)
	for i := range ports {
		ports[i] = i + 1
	}
	return ports
}

// parsePortList parses a comma-separated list of port numbers.
func parsePortList(s string) []int {
	if s == "" {
		return nil
	}
	var ports []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil && n > 0 && n <= 65535 {
			ports = append(ports, n)
		}
	}
	return ports
}

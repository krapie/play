package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	validTarget = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.\-]*$`)
	hopRe       = regexp.MustCompile(`^\s*(\d+)\s+(\*|[\d.]+)\s*(?:([\d.]+)\s+ms)?`)
)

type hopEvent struct {
	Hop     int     `json:"hop"`
	IP      string  `json:"ip,omitempty"`
	RTT     string  `json:"rtt,omitempty"`
	City    string  `json:"city,omitempty"`
	Country string  `json:"country,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
	AS      string  `json:"as,omitempty"`
	Timeout bool    `json:"timeout,omitempty"`
}

type ipAPIResp struct {
	Status      string  `json:"status"`
	Query       string  `json:"query"`
	City        string  `json:"city"`
	CountryCode string  `json:"countryCode"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AS          string  `json:"as"`
}

func (h *Handler) Tracer(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	if target == "" || len(target) > 253 || !validTarget.MatchString(target) {
		http.Error(w, "invalid target", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	geoClient := &http.Client{Timeout: 3 * time.Second}

	// The trace starts at the server's public location (private hops are dropped).
	if o := serverOrigin(geoClient); o != nil {
		data, _ := json.Marshal(hopEvent{IP: o.Query, City: o.City, Country: o.CountryCode, Lat: o.Lat, Lon: o.Lon, AS: o.AS})
		fmt.Fprintf(w, "event: origin\ndata: %s\n\n", data)
		flusher.Flush()
	}

	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "traceroute", "-4", "-n", "-q", "1", "-w", "2", "-m", "30", target)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"msg\":%q}\n\n", err.Error())
		flusher.Flush()
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"msg\":%q}\n\n", err.Error())
		flusher.Flush()
		return
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			return
		default:
		}

		line := scanner.Text()
		if strings.Contains(line, "traceroute to") {
			continue
		}

		m := hopRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		hop := &hopEvent{}
		fmt.Sscan(m[1], &hop.Hop)

		if m[2] == "*" {
			hop.Timeout = true
		} else {
			// Private hops (pod gateway, home router, CGNAT) can't be placed on
			// the map and only expose internal topology, so they are not sent.
			if isPrivate(m[2]) {
				continue
			}
			hop.IP = m[2]
			if len(m) > 3 && m[3] != "" {
				hop.RTT = m[3]
			}
			if geo := geoLookup(geoClient, hop.IP); geo != nil {
				hop.City = geo.City
				hop.Country = geo.CountryCode
				hop.Lat = geo.Lat
				hop.Lon = geo.Lon
				hop.AS = geo.AS
			}
		}

		data, _ := json.Marshal(hop)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	cmd.Wait()
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

var (
	originMu sync.Mutex
	origin   *ipAPIResp
)

// serverOrigin geolocates this server's public IP once (ip-api with no IP
// answers for the caller) and caches it; a failed lookup is retried next time.
func serverOrigin(client *http.Client) *ipAPIResp {
	originMu.Lock()
	defer originMu.Unlock()
	if origin == nil {
		origin = geoLookup(client, "")
	}
	return origin
}

func geoLookup(client *http.Client, ip string) *ipAPIResp {
	resp, err := client.Get("http://ip-api.com/json/" + ip + "?fields=status,query,city,countryCode,lat,lon,as")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var geo ipAPIResp
	if json.NewDecoder(resp.Body).Decode(&geo) != nil || geo.Status != "success" {
		return nil
	}
	return &geo
}

var privateNets []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10",
		"127.0.0.0/8", "169.254.0.0/16", "::1/128", "fc00::/7",
	} {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil {
			privateNets = append(privateNets, block)
		}
	}
}

func isPrivate(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}
	for _, block := range privateNets {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

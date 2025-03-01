package dns

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// DNSListener implements the Listener interface for DNS protocol
type DNSListener struct {
	config     DNSConfig
	server     *dns.Server
	status     string
	statusLock sync.RWMutex
	stopChan   chan struct{}
	ttlCache   map[string]time.Time
	cacheLock  sync.RWMutex
}

// DNSConfig holds configuration for the DNS listener
type DNSConfig struct {
	Address     string
	Port        int
	Domain      string
	TTL         uint32
	RandomDelay struct {
		Min time.Duration
		Max time.Duration
	}
	Options map[string]interface{}
}

// NewDNSListener creates a new DNS listener
func NewDNSListener(config DNSConfig) *DNSListener {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Set default values if not provided
	if config.TTL == 0 {
		config.TTL = 60 // Default TTL of 60 seconds
	}
	if config.RandomDelay.Min == 0 {
		config.RandomDelay.Min = 100 * time.Millisecond
	}
	if config.RandomDelay.Max == 0 {
		config.RandomDelay.Max = 500 * time.Millisecond
	}

	return &DNSListener{
		config:    config,
		status:    "stopped",
		stopChan:  make(chan struct{}),
		ttlCache:  make(map[string]time.Time),
	}
}

// Start implements the Listener interface
func (l *DNSListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("DNS listener is already running")
	}

	// Create DNS server
	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	l.server = &dns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: dns.HandlerFunc(l.handleDNSRequest),
	}

	// Start DNS server in a goroutine
	go func() {
		err := l.server.ListenAndServe()
		if err != nil {
			l.statusLock.Lock()
			l.status = "error"
			l.statusLock.Unlock()
			fmt.Printf("Error starting DNS listener: %v\n", err)
		}
	}()

	// Start cache cleanup routine
	go l.cleanupCache()

	l.status = "running"
	return nil
}

// Stop implements the Listener interface
func (l *DNSListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != "running" {
		return nil // Already stopped
	}

	// Signal the cleanup goroutine to stop
	close(l.stopChan)

	// Create a new stop channel for future use
	l.stopChan = make(chan struct{})

	// Stop the DNS server
	if l.server != nil {
		err := l.server.Shutdown()
		if err != nil {
			l.status = "error"
			return fmt.Errorf("error stopping DNS listener: %w", err)
		}
	}

	l.status = "stopped"
	return nil
}

// Status implements the Listener interface
func (l *DNSListener) Status() string {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *DNSListener) Configure(config interface{}) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("cannot configure a running DNS listener")
	}

	dnsConfig, ok := config.(DNSConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type for DNS listener")
	}

	l.config = dnsConfig
	return nil
}

// handleDNSRequest processes incoming DNS requests
func (l *DNSListener) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	// Process each question
	for _, q := range r.Question {
		// Check if the query is for our domain
		if dns.IsSubDomain(l.config.Domain+".", q.Name) {
			// Extract the subdomain part (which contains our data)
			subdomain := l.extractSubdomain(q.Name)

			// Check cache to avoid duplicate processing
			l.cacheLock.RLock()
			lastSeen, exists := l.ttlCache[subdomain]
			l.cacheLock.RUnlock()

			if exists && time.Since(lastSeen) < time.Duration(l.config.TTL)*time.Second {
				// This is a cached query, likely from a recursive resolver
				// Just respond with the same data
				rr := l.createResponseRecord(q, subdomain)
				m.Answer = append(m.Answer, rr)
			} else {
				// This is a new query or the cache has expired
				// Process the data in the subdomain
				data := l.decodeSubdomain(subdomain)

				// Add to cache
				l.cacheLock.Lock()
				l.ttlCache[subdomain] = time.Now()
				l.cacheLock.Unlock()

				// Create response
				rr := l.createResponseRecord(q, data)
				m.Answer = append(m.Answer, rr)

				// Process the actual data (in a real implementation, this would
				// be passed to the protocol layer)
				go l.processData(data, w.RemoteAddr())
			}
		}
	}

	// Add random delay to avoid detection
	delay := l.randomDelay()
	time.Sleep(delay)

	// Send response
	w.WriteMsg(m)
}

// extractSubdomain extracts the subdomain part from a DNS query
func (l *DNSListener) extractSubdomain(name string) string {
	// Remove the base domain and trailing dot
	return name[:len(name)-len(l.config.Domain)-2]
}

// decodeSubdomain decodes the data encoded in the subdomain
func (l *DNSListener) decodeSubdomain(subdomain string) string {
	// In a real implementation, this would decode the data
	// For now, just return the subdomain as-is
	return subdomain
}

// createResponseRecord creates a DNS response record
func (l *DNSListener) createResponseRecord(q dns.Question, data string) dns.RR {
	// In a real implementation, this would encode the response data
	// For now, just create a simple TXT record
	txt := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   q.Name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    l.config.TTL,
		},
		Txt: []string{data},
	}
	return txt
}

// processData processes the data received in a DNS query
func (l *DNSListener) processData(data string, addr net.Addr) {
	// In a real implementation, this would pass the data to the protocol layer
	fmt.Printf("Received data from %s: %s\n", addr, data)
}

// randomDelay returns a random delay within the configured range
func (l *DNSListener) randomDelay() time.Duration {
	min := int64(l.config.RandomDelay.Min)
	max := int64(l.config.RandomDelay.Max)
	return time.Duration(rand.Int63n(max-min) + min)
}

// cleanupCache periodically cleans up expired entries from the cache
func (l *DNSListener) cleanupCache() {
	ticker := time.NewTicker(time.Duration(l.config.TTL) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cacheLock.Lock()
			now := time.Now()
			for subdomain, lastSeen := range l.ttlCache {
				if now.Sub(lastSeen) > time.Duration(l.config.TTL)*time.Second {
					delete(l.ttlCache, subdomain)
				}
			}
			l.cacheLock.Unlock()
		case <-l.stopChan:
			return
		}
	}
}

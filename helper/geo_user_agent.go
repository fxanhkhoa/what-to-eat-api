package helper

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GeoInfo struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Status      string `json:"status"`
}

var (
	geoCache sync.Map // map[string]*GeoInfo
	httpClient = &http.Client{
		Timeout: 2 * time.Second,
	}
)

// ParseUserAgent extracts Browser, OS, and Device from a User-Agent string
func ParseUserAgent(ua string) (browser string, os string, device string) {
	if ua == "" {
		return "Unknown", "Unknown", "Unknown"
	}

	lower := strings.ToLower(ua)

	// Device detection
	if strings.Contains(lower, "ipad") || (strings.Contains(lower, "tablet") || (strings.Contains(lower, "android") && !strings.Contains(lower, "mobile"))) {
		device = "Tablet"
	} else if strings.Contains(lower, "mobile") || strings.Contains(lower, "iphone") || strings.Contains(lower, "ipod") || strings.Contains(lower, "android") {
		device = "Mobile"
	} else {
		device = "Desktop"
	}

	// OS detection
	switch {
	case strings.Contains(lower, "windows phone"):
		os = "Windows Phone"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ipod"):
		os = "iOS"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x"):
		os = "macOS"
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "cros"):
		os = "Chrome OS"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	default:
		os = "Other"
	}

	// Browser detection (order matters because Chromium browsers include "Chrome", "Safari", etc.)
	switch {
	case strings.Contains(lower, "edg/") || strings.Contains(lower, "edge/"):
		browser = "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera/"):
		browser = "Opera"
	case strings.Contains(lower, "chrome/") || strings.Contains(lower, "crios/"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/") || strings.Contains(lower, "fxios/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome"):
		browser = "Safari"
	default:
		browser = "Other"
	}

	return browser, os, device
}

// IsPrivateOrLoopbackIP checks if the IP address is private, loopback, or invalid
func IsPrivateOrLoopbackIP(ipStr string) bool {
	cleanIP := strings.TrimSpace(ipStr)
	// Handle IP with port
	if host, _, err := net.SplitHostPort(cleanIP); err == nil {
		cleanIP = host
	}

	ip := net.ParseIP(cleanIP)
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// ResolveGeoIP resolves IP into Country, CountryCode, and City
func ResolveGeoIP(ipStr string) *GeoInfo {
	cleanIP := strings.TrimSpace(ipStr)
	if host, _, err := net.SplitHostPort(cleanIP); err == nil {
		cleanIP = host
	}

	if IsPrivateOrLoopbackIP(cleanIP) {
		return &GeoInfo{
			Country:     "Local",
			CountryCode: "LOC",
			RegionName:  "Local",
			City:        "Local",
			Status:      "success",
		}
	}

	if val, ok := geoCache.Load(cleanIP); ok {
		if info, ok := val.(*GeoInfo); ok {
			return info
		}
	}

	resp, err := httpClient.Get("http://ip-api.com/json/" + cleanIP + "?fields=status,country,countryCode,regionName,city")
	if err != nil {
		return &GeoInfo{
			Country:     "Unknown",
			CountryCode: "UN",
			Status:      "fail",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &GeoInfo{
			Country:     "Unknown",
			CountryCode: "UN",
			Status:      "fail",
		}
	}

	var info GeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Status != "success" {
		return &GeoInfo{
			Country:     "Unknown",
			CountryCode: "UN",
			Status:      "fail",
		}
	}

	geoCache.Store(cleanIP, &info)
	return &info
}


package helper

import (
	"testing"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name       string
		ua         string
		expBrowser string
		expOS      string
		expDevice  string
	}{
		{
			name:       "Empty UA",
			ua:         "",
			expBrowser: "Unknown",
			expOS:      "Unknown",
			expDevice:  "Unknown",
		},
		{
			name:       "Chrome on macOS Desktop",
			ua:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expBrowser: "Chrome",
			expOS:      "macOS",
			expDevice:  "Desktop",
		},
		{
			name:       "Safari on iPhone Mobile",
			ua:         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			expBrowser: "Safari",
			expOS:      "iOS",
			expDevice:  "Mobile",
		},
		{
			name:       "Firefox on Windows Desktop",
			ua:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0",
			expBrowser: "Firefox",
			expOS:      "Windows",
			expDevice:  "Desktop",
		},
		{
			name:       "Edge on Windows Desktop",
			ua:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
			expBrowser: "Edge",
			expOS:      "Windows",
			expDevice:  "Desktop",
		},
		{
			name:       "Chrome on Android Tablet",
			ua:         "Mozilla/5.0 (Linux; Android 12; SM-X906N Build/SP1A.210812.016) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
			expBrowser: "Chrome",
			expOS:      "Android",
			expDevice:  "Tablet",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, o, d := ParseUserAgent(tc.ua)
			if b != tc.expBrowser {
				t.Errorf("expected browser %s, got %s", tc.expBrowser, b)
			}
			if o != tc.expOS {
				t.Errorf("expected OS %s, got %s", tc.expOS, o)
			}
			if d != tc.expDevice {
				t.Errorf("expected device %s, got %s", tc.expDevice, d)
			}
		})
	}
}

func TestIsPrivateOrLoopbackIP(t *testing.T) {
	if !IsPrivateOrLoopbackIP("127.0.0.1") {
		t.Errorf("expected 127.0.0.1 to be private/loopback")
	}
	if !IsPrivateOrLoopbackIP("::1") {
		t.Errorf("expected ::1 to be private/loopback")
	}
	if !IsPrivateOrLoopbackIP("192.168.1.1") {
		t.Errorf("expected 192.168.1.1 to be private/loopback")
	}
	if !IsPrivateOrLoopbackIP("10.0.0.5:8080") {
		t.Errorf("expected 10.0.0.5:8080 to be private/loopback")
	}
	if IsPrivateOrLoopbackIP("8.8.8.8") {
		t.Errorf("expected 8.8.8.8 to NOT be private/loopback")
	}
}

func TestResolveGeoIPLocal(t *testing.T) {
	geo := ResolveGeoIP("127.0.0.1")
	if geo == nil || geo.Country != "Local" {
		t.Errorf("expected Local country for 127.0.0.1, got %+v", geo)
	}
}


package config

import "testing"

func TestMCPHTTPConfig_ResolveListen_bindLocal(t *testing.T) {
	c := MCPHTTPConfig{Listen: ":8091", BindLocal: true}
	if got := c.ResolveListen(); got != "127.0.0.1:8091" {
		t.Fatalf("got %q", got)
	}
}

func TestMCPHTTPConfig_ResolveListen_default(t *testing.T) {
	c := MCPHTTPConfig{Listen: ":9090"}
	if got := c.ResolveListen(); got != ":9090" {
		t.Fatalf("got %q", got)
	}
}

func TestMCPHTTPConfig_ResolveListen_empty(t *testing.T) {
	c := MCPHTTPConfig{}
	if got := c.ResolveListen(); got != ":8091" {
		t.Fatalf("got %q", got)
	}
}

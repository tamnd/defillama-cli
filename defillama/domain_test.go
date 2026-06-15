package defillama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "defillama" {
		t.Errorf("Scheme = %q, want defillama", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "defillama" {
		t.Errorf("Identity.Binary = %q, want defillama", info.Identity.Binary)
	}
}

func TestListProtocols(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/protocols" {
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":       "1",
					"name":     "Uniswap",
					"symbol":   "UNI",
					"chain":    "Ethereum",
					"category": "DEX",
					"tvl":      5000000000.0,
					"change_1d": 1.5,
					"change_7d": -2.3,
					"url":      "https://uniswap.org",
				},
				{
					"id":       "2",
					"name":     "Aave",
					"symbol":   "AAVE",
					"chain":    "Ethereum",
					"category": "Lending",
					"tvl":      8000000000.0,
					"change_1d": 0.5,
					"change_7d": 1.2,
					"url":      "https://aave.com",
				},
			})
		}
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	// No filter — should return both sorted by TVL descending.
	protocols, err := c.ListProtocols(context.Background(), "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(protocols) != 2 {
		t.Fatalf("got %d protocols, want 2", len(protocols))
	}
	if protocols[0].Name != "Aave" {
		t.Errorf("first protocol = %q, want Aave (highest TVL)", protocols[0].Name)
	}
	if protocols[1].Name != "Uniswap" {
		t.Errorf("second protocol = %q, want Uniswap", protocols[1].Name)
	}
}

func TestListProtocolsFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"name": "Uniswap", "symbol": "UNI", "chain": "Ethereum", "category": "DEX", "tvl": 5e9},
			{"name": "Aave", "symbol": "AAVE", "chain": "Ethereum", "category": "Lending", "tvl": 8e9},
			{"name": "PancakeSwap", "symbol": "CAKE", "chain": "BSC", "category": "DEX", "tvl": 3e9},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	// Filter by category=DEX.
	protocols, err := c.ListProtocols(context.Background(), "DEX", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(protocols) != 2 {
		t.Fatalf("DEX filter: got %d protocols, want 2", len(protocols))
	}

	// Filter by chain=BSC.
	protocols, err = c.ListProtocols(context.Background(), "", "BSC", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(protocols) != 1 || protocols[0].Name != "PancakeSwap" {
		t.Errorf("BSC filter: got %v, want [PancakeSwap]", protocols)
	}

	// Limit.
	protocols, err = c.ListProtocols(context.Background(), "", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(protocols) != 1 {
		t.Errorf("limit=1: got %d protocols, want 1", len(protocols))
	}
}

func TestListChains(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"name": "Ethereum", "tokenSymbol": "ETH", "tvl": 50000000000.0},
			{"name": "BSC", "tokenSymbol": "BNB", "tvl": 5000000000.0},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	chains, err := c.ListChains(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 2 {
		t.Fatalf("got %d chains, want 2", len(chains))
	}
	if chains[0].Name != "Ethereum" {
		t.Errorf("first chain = %q, want Ethereum", chains[0].Name)
	}
	if chains[0].Symbol != "ETH" {
		t.Errorf("chain symbol = %q, want ETH", chains[0].Symbol)
	}
}

func TestListStablecoins(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"peggedAssets": []map[string]interface{}{
				{
					"name":         "Tether",
					"symbol":       "USDT",
					"pegType":      "peggedUSD",
					"pegMechanism": "fiat-backed",
					"price":        1.0,
					"circulating": map[string]interface{}{
						"peggedUSD": 120000000000.0,
					},
				},
				{
					"name":         "USD Coin",
					"symbol":       "USDC",
					"pegType":      "peggedUSD",
					"pegMechanism": "fiat-backed",
					"price":        1.0,
					"circulating": map[string]interface{}{
						"peggedUSD": 40000000000.0,
					},
				},
			},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	stables, err := c.ListStablecoins(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(stables) != 2 {
		t.Fatalf("got %d stablecoins, want 2", len(stables))
	}
	if stables[0].Name != "Tether" {
		t.Errorf("first stablecoin = %q, want Tether (highest circ)", stables[0].Name)
	}
	if stables[0].CircUSD != 120000000000.0 {
		t.Errorf("CircUSD = %v, want 120000000000", stables[0].CircUSD)
	}
}

func TestListYields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"data": []map[string]interface{}{
				{
					"chain":      "Ethereum",
					"project":    "aave-v3",
					"symbol":     "USDC",
					"tvlUsd":     500000000.0,
					"apy":        3.5,
					"pool":       "pool-1",
					"stableCoin": true,
				},
				{
					"chain":      "BSC",
					"project":    "pancakeswap",
					"symbol":     "CAKE-BNB",
					"tvlUsd":     100000000.0,
					"apy":        25.0,
					"pool":       "pool-2",
					"stableCoin": false,
				},
			},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	// No filter.
	pools, err := c.ListYields(context.Background(), 0, "", "", false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 2 {
		t.Fatalf("got %d pools, want 2", len(pools))
	}
	// Sorted by TVL descending: aave-v3 first.
	if pools[0].Project != "aave-v3" {
		t.Errorf("first pool project = %q, want aave-v3", pools[0].Project)
	}

	// Filter by min APY.
	pools, err = c.ListYields(context.Background(), 10.0, "", "", false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 || pools[0].Project != "pancakeswap" {
		t.Errorf("min-apy=10 filter: got %v, want [pancakeswap]", pools)
	}

	// Filter stablecoin only.
	pools, err = c.ListYields(context.Background(), 0, "", "", true, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 || !pools[0].Stablecoin {
		t.Errorf("stablecoin filter: got %v", pools)
	}
}

func TestGetProtocol(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/protocol/uniswap" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":      "Uniswap",
				"symbol":    "UNI",
				"chain":     "Ethereum",
				"category":  "DEX",
				"tvl":       5000000000.0,
				"change_1d": 1.5,
				"change_7d": -2.3,
				"url":       "https://uniswap.org",
			})
		}
	}))
	defer ts.Close()

	c := NewClient()
	c.Rate = 0
	c.baseURL = ts.URL
	c.stablecoinsURL = ts.URL
	c.yieldsURL = ts.URL

	p, err := c.GetProtocol(context.Background(), "uniswap")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Uniswap" {
		t.Errorf("Name = %q, want Uniswap", p.Name)
	}
	if p.TVL != 5000000000.0 {
		t.Errorf("TVL = %v, want 5000000000", p.TVL)
	}
	if p.Category != "DEX" {
		t.Errorf("Category = %q, want DEX", p.Category)
	}
}

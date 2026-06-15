// Package defillama is the library behind the defillama command line:
// the HTTP client, request shaping, and the typed data models for the
// DeFi Llama API.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite, and retries the
// transient failures (429 and 5xx) that any public API throws under load.
package defillama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// Host is the primary API host.
const Host = "api.llama.fi"

// BaseURL is the root every main request is built from.
const BaseURL = "https://" + Host

// StablecoinsURL is the stablecoins subdomain.
const StablecoinsURL = "https://stablecoins.llama.fi"

// YieldsURL is the yields subdomain.
const YieldsURL = "https://yields.llama.fi"

// DefaultUserAgent identifies the client to DeFi Llama.
const DefaultUserAgent = "defillama-cli/0.1 (tamnd87@gmail.com)"

// Config holds the client configuration.
type Config struct {
	BaseURL        string
	StablecoinsURL string
	YieldsURL      string
	UserAgent      string
	Rate           time.Duration
	Timeout        time.Duration
	Retries        int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:        BaseURL,
		StablecoinsURL: StablecoinsURL,
		YieldsURL:      YieldsURL,
		Rate:           500 * time.Millisecond,
		Timeout:        30 * time.Second,
		Retries:        3,
		UserAgent:      DefaultUserAgent,
	}
}

// Client talks to the DeFi Llama API over HTTP.
type Client struct {
	HTTP           *http.Client
	UserAgent      string
	Rate           time.Duration
	Retries        int
	baseURL        string
	stablecoinsURL string
	yieldsURL      string
	last           time.Time
}

// NewClient returns a Client with sensible defaults.
func NewClient() *Client {
	cfg := DefaultConfig()
	return &Client{
		HTTP:           &http.Client{Timeout: cfg.Timeout},
		UserAgent:      cfg.UserAgent,
		Rate:           cfg.Rate,
		Retries:        cfg.Retries,
		baseURL:        cfg.BaseURL,
		stablecoinsURL: cfg.StablecoinsURL,
		yieldsURL:      cfg.YieldsURL,
	}
}

// NewClientFromConfig returns a Client configured from cfg.
func NewClientFromConfig(cfg Config) *Client {
	c := NewClient()
	if cfg.BaseURL != "" {
		c.baseURL = cfg.BaseURL
	}
	if cfg.StablecoinsURL != "" {
		c.stablecoinsURL = cfg.StablecoinsURL
	}
	if cfg.YieldsURL != "" {
		c.yieldsURL = cfg.YieldsURL
	}
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	return c
}

// Get fetches a URL and returns the response body.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// --- output types ---

// Protocol is a DeFi protocol with TVL data.
type Protocol struct {
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Chain    string  `json:"chain"`
	Category string  `json:"category"`
	TVL      float64 `json:"tvl"`
	Change1D float64 `json:"change_1d"`
	Change7D float64 `json:"change_7d"`
	URL      string  `json:"url"`
}

// Chain is a blockchain network.
type Chain struct {
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	TVL    float64 `json:"tvl"`
}

// Stablecoin is a pegged digital asset.
type Stablecoin struct {
	Name      string  `json:"name"`
	Symbol    string  `json:"symbol"`
	PegType   string  `json:"peg_type"`
	Mechanism string  `json:"mechanism"`
	CircUSD   float64 `json:"circulating_usd"`
	Price     float64 `json:"price"`
}

// YieldPool is a DeFi yield farming pool.
type YieldPool struct {
	Chain      string  `json:"chain"`
	Project    string  `json:"project"`
	Symbol     string  `json:"symbol"`
	TVLUsd     float64 `json:"tvl_usd"`
	APY        float64 `json:"apy"`
	PoolID     string  `json:"pool_id"`
	Stablecoin bool    `json:"stablecoin"`
}

// --- wire types ---

type wireProtocol struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Symbol   string   `json:"symbol"`
	Chain    string   `json:"chain"`
	Logo     string   `json:"logo"`
	TVL      float64  `json:"tvl"`
	Change1H float64  `json:"change_1h"`
	Change1D float64  `json:"change_1d"`
	Change7D float64  `json:"change_7d"`
	Category string   `json:"category"`
	Chains   []string `json:"chains"`
	URL      string   `json:"url"`
}

type wireChain struct {
	Name        string  `json:"name"`
	GeckoID     string  `json:"gecko_id"`
	TVL         float64 `json:"tvl"`
	TokenSymbol string  `json:"tokenSymbol"`
	CMCID       int     `json:"cmcId"`
	ChainID     int     `json:"chainId"`
}

type wireStablecoinsResp struct {
	PeggedAssets []wireStablecoin `json:"peggedAssets"`
}

type wireStablecoin struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Symbol       string                 `json:"symbol"`
	PegType      string                 `json:"pegType"`
	PegMechanism string                 `json:"pegMechanism"`
	Price        float64                `json:"price"`
	Circulating  map[string]interface{} `json:"circulating"`
}

type wirePoolsResp struct {
	Status string     `json:"status"`
	Data   []wirePool `json:"data"`
}

type wirePool struct {
	Chain      string  `json:"chain"`
	Project    string  `json:"project"`
	Symbol     string  `json:"symbol"`
	TVLUsd     float64 `json:"tvlUsd"`
	APY        float64 `json:"apy"`
	APYBase    float64 `json:"apyBase"`
	APYReward  float64 `json:"apyReward"`
	Pool       string  `json:"pool"`
	StableCoin bool    `json:"stableCoin"`
	ILRisk     string  `json:"ilRisk"`
}

// --- API methods ---

// ListProtocols fetches all protocols and applies optional filters.
func (c *Client) ListProtocols(ctx context.Context, category, chain string, limit int) ([]*Protocol, error) {
	body, err := c.Get(ctx, c.baseURL+"/protocols")
	if err != nil {
		return nil, err
	}

	var raw []wireProtocol
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode protocols: %w", err)
	}

	var out []*Protocol
	for _, w := range raw {
		if category != "" && w.Category != category {
			continue
		}
		if chain != "" && w.Chain != chain {
			continue
		}
		out = append(out, &Protocol{
			Name:     w.Name,
			Symbol:   w.Symbol,
			Chain:    w.Chain,
			Category: w.Category,
			TVL:      w.TVL,
			Change1D: w.Change1D,
			Change7D: w.Change7D,
			URL:      w.URL,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].TVL > out[j].TVL
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// GetProtocol fetches a single protocol by slug.
func (c *Client) GetProtocol(ctx context.Context, slug string) (*Protocol, error) {
	body, err := c.Get(ctx, c.baseURL+"/protocol/"+slug)
	if err != nil {
		return nil, err
	}

	var w wireProtocol
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, fmt.Errorf("decode protocol: %w", err)
	}

	return &Protocol{
		Name:     w.Name,
		Symbol:   w.Symbol,
		Chain:    w.Chain,
		Category: w.Category,
		TVL:      w.TVL,
		Change1D: w.Change1D,
		Change7D: w.Change7D,
		URL:      w.URL,
	}, nil
}

// ListChains fetches all chains sorted by TVL.
func (c *Client) ListChains(ctx context.Context, limit int) ([]*Chain, error) {
	body, err := c.Get(ctx, c.baseURL+"/v2/chains")
	if err != nil {
		return nil, err
	}

	var raw []wireChain
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode chains: %w", err)
	}

	out := make([]*Chain, len(raw))
	for i, w := range raw {
		out[i] = &Chain{
			Name:   w.Name,
			Symbol: w.TokenSymbol,
			TVL:    w.TVL,
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].TVL > out[j].TVL
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ListStablecoins fetches all stablecoins sorted by circulating USD.
func (c *Client) ListStablecoins(ctx context.Context, limit int) ([]*Stablecoin, error) {
	body, err := c.Get(ctx, c.stablecoinsURL+"/stablecoins?includePrices=true")
	if err != nil {
		return nil, err
	}

	var resp wireStablecoinsResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode stablecoins: %w", err)
	}

	var out []*Stablecoin
	for _, w := range resp.PeggedAssets {
		var circUSD float64
		if w.Circulating != nil {
			if v, ok := w.Circulating["peggedUSD"]; ok {
				switch n := v.(type) {
				case float64:
					circUSD = n
				}
			}
		}
		out = append(out, &Stablecoin{
			Name:      w.Name,
			Symbol:    w.Symbol,
			PegType:   w.PegType,
			Mechanism: w.PegMechanism,
			CircUSD:   circUSD,
			Price:     w.Price,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CircUSD > out[j].CircUSD
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ListYields fetches yield pools with optional filters.
func (c *Client) ListYields(ctx context.Context, minAPY float64, chain, project string, stablecoinOnly bool, limit int) ([]*YieldPool, error) {
	body, err := c.Get(ctx, c.yieldsURL+"/pools")
	if err != nil {
		return nil, err
	}

	var resp wirePoolsResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode pools: %w", err)
	}

	var out []*YieldPool
	for _, w := range resp.Data {
		if minAPY > 0 && w.APY < minAPY {
			continue
		}
		if chain != "" && w.Chain != chain {
			continue
		}
		if project != "" && w.Project != project {
			continue
		}
		if stablecoinOnly && !w.StableCoin {
			continue
		}
		out = append(out, &YieldPool{
			Chain:      w.Chain,
			Project:    w.Project,
			Symbol:     w.Symbol,
			TVLUsd:     w.TVLUsd,
			APY:        w.APY,
			PoolID:     w.Pool,
			Stablecoin: w.StableCoin,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].TVLUsd > out[j].TVLUsd
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

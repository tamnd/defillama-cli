package defillama

import (
	"context"

	"github.com/tamnd/any-cli/kit"
)

// domain.go exposes DeFi Llama as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/defillama-cli/defillama"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// defillama:// URIs by routing to the operations Register installs. The standalone
// defillama binary does not use any of this, so the CLI is unchanged.
func init() { kit.Register(Domain{}) }

// Domain is the DeFi Llama driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity a host reuses for help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "defillama",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "defillama",
			Short:  "DeFi Llama: 7,661+ DeFi protocols, 452 chains, yield pools",
			Long: `DeFi Llama is the most comprehensive DeFi data aggregator.

defillama reads public data over plain HTTPS, shapes it into clean records,
and prints output that pipes into the rest of your tools. No API key, nothing
to run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/defillama-cli",
		},
	}
}

// Register installs the client factory and every DeFi Llama operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "protocols", Group: "read", List: true,
		Summary: "List DeFi protocols by TVL (--category, --chain, --limit)"}, listProtocols)

	kit.Handle(app, kit.OpMeta{Name: "chains", Group: "read", List: true,
		Summary: "List blockchain chains by TVL (--limit)"}, listChains)

	kit.Handle(app, kit.OpMeta{Name: "stablecoins", Group: "read", List: true,
		Summary: "List stablecoins by market cap (--limit)"}, listStablecoins)

	kit.Handle(app, kit.OpMeta{Name: "yields", Group: "read", List: true,
		Summary: "List yield farming pools (--min-apy, --chain, --project, --stablecoin)"}, listYields)

	kit.Handle(app, kit.OpMeta{Name: "tvl", Group: "read", Single: true,
		Summary: "Get TVL data for a specific protocol",
		Args:    []kit.Arg{{Name: "slug", Help: "protocol slug (e.g. uniswap, aave)"}}}, getProtocolTVL)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// --- inputs ---

type protocolsInput struct {
	Category string  `kit:"flag" help:"filter by category (DEX, Lending, Bridge, etc)"`
	Chain    string  `kit:"flag" help:"filter by chain (Ethereum, BSC, etc)"`
	Limit    int     `kit:"flag,inherit" help:"max results"`
	Client   *Client `kit:"inject"`
}

type chainsInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type stablecoinsInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type yieldsInput struct {
	MinAPY     float64 `kit:"flag" help:"minimum APY filter"`
	Chain      string  `kit:"flag" help:"chain filter"`
	Project    string  `kit:"flag" help:"project filter"`
	Stablecoin bool    `kit:"flag" help:"filter to stablecoin pools only"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

type tvlInput struct {
	Slug   string  `kit:"arg" help:"protocol slug"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listProtocols(ctx context.Context, in protocolsInput, emit func(*Protocol) error) error {
	protocols, err := in.Client.ListProtocols(ctx, in.Category, in.Chain, in.Limit)
	if err != nil {
		return err
	}
	for _, p := range protocols {
		if err := emit(p); err != nil {
			return err
		}
	}
	return nil
}

func listChains(ctx context.Context, in chainsInput, emit func(*Chain) error) error {
	chains, err := in.Client.ListChains(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, ch := range chains {
		if err := emit(ch); err != nil {
			return err
		}
	}
	return nil
}

func listStablecoins(ctx context.Context, in stablecoinsInput, emit func(*Stablecoin) error) error {
	stables, err := in.Client.ListStablecoins(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, s := range stables {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func listYields(ctx context.Context, in yieldsInput, emit func(*YieldPool) error) error {
	pools, err := in.Client.ListYields(ctx, in.MinAPY, in.Chain, in.Project, in.Stablecoin, in.Limit)
	if err != nil {
		return err
	}
	for _, p := range pools {
		if err := emit(p); err != nil {
			return err
		}
	}
	return nil
}

func getProtocolTVL(ctx context.Context, in tvlInput, emit func(*Protocol) error) error {
	p, err := in.Client.GetProtocol(ctx, in.Slug)
	if err != nil {
		return err
	}
	return emit(p)
}

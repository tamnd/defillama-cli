---
title: "defillama"
description: "DeFi Llama: 7,661+ DeFi protocols, 452 chains, yield pools"
heroTitle: "defillama, from the command line"
heroLead: "DeFi Llama: 7,661+ DeFi protocols, 452 chains, yield pools One pure-Go binary, no API key, output that pipes into the rest of your tools, and a resource-URI driver other programs can address."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`defillama` reads public defillama data over plain HTTPS, shapes it into
clean records, and gets out of your way.

```bash
defillama page <path>            # fetch one page as a record
defillama page <path> -o json    # as JSON, ready for jq
defillama links <path>           # the pages it links to, each addressable
defillama serve --addr :7777     # the same operations over HTTP
```

There is nothing to sign up for and nothing to run alongside it. Output adapts
to where it goes: an aligned table on your terminal, JSONL the moment you pipe
it somewhere.

## Two ways to use it

- **As a command** for reading defillama by hand or in a script. Start with
  the [quick start](/getting-started/quick-start/).
- **As a resource-URI driver** so a host like
  [ant](https://github.com/tamnd/ant) can address defillama as
  `defillama://` URIs and follow links across sites. See
  [resource URIs](/guides/resource-uris/).

Both are the same code: one operation, declared once, is a CLI command, an HTTP
route, an MCP tool, and a URI dereference.

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Doing a specific job? The [guides](/guides/) are task-first.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.

# Monitor types

Ogoune ships several check strategies.

| Type | Checks |
|---|---|
| **HTTP** | Status code, response time, redirects |
| **TCP** | Port reachability |
| **DNS** | Record resolution |
| **ICMP** | Ping / host reachability |
| **Keyword** | Presence/absence of a string in the response body |
| **Protocol** | Protocol-specific handshakes |

Each strategy implements a common `CheckStrategy` contract, so new types can be added without touching the scheduler or worker pool.

---
layout: home

hero:
  name: Ogoune
  text: Uptime monitoring that confirms before it cries wolf.
  tagline: Verifies failures before alerting. No more 3am pages for a 2-second network blip.
  image:
    src: /dashboard.png
    alt: Ogoune dashboard
  actions:
    - theme: brand
      text: Get started
      link: /guide/
    - theme: alt
      text: Self-host
      link: /self-host/
    - theme: alt
      text: View on GitHub
      link: https://github.com/denisakp/ogoune

features:
  - title: Confirmed failures only
    details: Requires N consecutive failed checks before opening an incident. Flap detection and alert grouping built in.
  - title: Open-core, self-hostable
    details: Community Edition (Apache 2.0) runs as a single binary on SQLite. Enterprise adds Postgres + Redis for horizontal scale.
  - title: Many monitor types
    details: HTTP, TCP, DNS, ICMP, keyword, and protocol checks — with SMTP, Slack, Discord, Google Chat, Teams, and webhook alerts.
  - title: Managed Cloud
    details: Don't want to run it yourself? Ogoune Cloud is the managed version — same code, we host it.
---

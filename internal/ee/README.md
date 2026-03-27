# Enterprise Edition Boundary

This folder contains edition-detection primitives for Open Core behavior.

Current scope:
- `license.Get()` returns `community` or `enterprise`
- `license.IsEnterprise()` returns a boolean helper

Detection rule:
- `ENTERPRISE_LICENSE_KEY` values that start with `pg_ent_` are treated as enterprise.
- All other values are treated as community.

This package only exposes runtime edition metadata and does not gate backend behavior in this feature.

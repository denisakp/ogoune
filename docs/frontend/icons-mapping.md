# Icon mapping — Ant Design Icons → Iconify (lucide / heroicons)

**Source**: Spec 053 Slice 1 / PR-1 · T018-T019
**Purpose**: Pin every legacy `@ant-design/icons-vue` import currently used in `web/src/` to its replacement under the NuxtUI / Iconify icon system. PR-3 (shared components + AppLayout) uses this table as a mechanical swap checklist.

## Audit command (reproducible)

```sh
grep -rhoE "import \{[^}]+\} from ['\"]@ant-design/icons-vue['\"]" web/src/ \
  | sed -E "s/import \{//; s/\} from.*//" \
  | tr ',' '\n' | tr -d ' ' | grep -v '^$' | sort -u
```

As of 2026-06-02 this returns 18 unique icons across 21 files. The same command is the canonical audit anchor for SC-004.

## Mapping

Convention: prefer `lucide` (NuxtUI default collection). Fall back to `heroicons` when no like-for-like lucide equivalent exists at similar weight.

| Legacy import name       | Replacement                | Collection | First seen at                                            | Notes |
|--------------------------|----------------------------|------------|----------------------------------------------------------|-------|
| `ArrowLeftOutlined`      | `i-lucide-arrow-left`      | lucide     | `web/src/components/resources/ResourceForm.vue`          |       |
| `ArrowRightOutlined`     | `i-lucide-arrow-right`     | lucide     | `web/src/components/resources/ResourceForm.vue`          |       |
| `DashboardOutlined`      | `i-lucide-layout-dashboard`| lucide     | `web/src/App.vue` (initial chunk)                        | High-traffic icon; verify weight in PR-3 |
| `EditOutlined`           | `i-lucide-pencil`          | lucide     | `web/src/views/SettingsView.vue`                         |       |
| `EllipsisOutlined`       | `i-lucide-ellipsis`        | lucide     | `web/src/views/MonitorsView.vue`                         |       |
| `EyeOutlined`            | `i-lucide-eye`             | lucide     | `web/src/views/auth/LoginView.vue`                       |       |
| `FallOutlined`           | `i-lucide-trending-down`   | lucide     | `web/src/views/resources/ResourcePerformance.vue`        |       |
| `FolderOutlined`         | `i-lucide-folder`          | lucide     | `web/src/views/ComponentsView.vue`                       |       |
| `GlobalOutlined`         | `i-lucide-globe`           | lucide     | `web/src/views/status-page/StatusPageDetailView.vue`     |       |
| `HeartOutlined`          | `i-lucide-heart`           | lucide     | `web/src/components/FeedbackModal.vue`                   |       |
| `LockOutlined`           | `i-lucide-lock`            | lucide     | `web/src/views/auth/LoginView.vue`                       |       |
| `MailOutlined`           | `i-lucide-mail`            | lucide     | `web/src/views/auth/LoginView.vue`                       |       |
| `PauseOutlined`          | `i-lucide-pause`           | lucide     | `web/src/views/resources/ResourceView.vue`               |       |
| `PlusOutlined`           | `i-lucide-plus`            | lucide     | `web/src/views/MonitorsView.vue`                         |       |
| `RiseOutlined`           | `i-lucide-trending-up`     | lucide     | `web/src/views/resources/ResourcePerformance.vue`        |       |
| `SafetyOutlined`         | `i-lucide-shield-check`    | lucide     | `web/src/views/auth/Verify2FAView.vue`                   | `shield-check` matches the "safety/verified" semantic better than raw `shield` |
| `SaveOutlined`           | `i-lucide-save`            | lucide     | `web/src/views/settings/AccountSettingsView.vue`         |       |
| `SettingOutlined`        | `i-lucide-settings`        | lucide     | `web/src/views/SettingsView.vue`                         |       |

**Coverage**: 18/18 (100%) — SC-004 ✅.

## Usage post-PR-3

```vue
<!-- before -->
<template>
  <SettingOutlined />
</template>
<script setup lang="ts">
import { SettingOutlined } from '@ant-design/icons-vue'
</script>

<!-- after -->
<template>
  <UIcon name="i-lucide-settings" />
</template>
<!-- no import needed: NuxtUI auto-resolves via Iconify -->
```

## Notes

- Iconify resolves names statically at build time when used as string literals; dynamic names defeat tree-shaking.
- The 21 import sites listed by the audit will be migrated in PR-3 alongside the AppLayout shell.
- The legacy `@ant-design/icons-vue` dep is dropped in Slice 6 (PRD 009) once all imports are gone — this table is the gate.

<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import type { PublicIncidentUpdate } from '@/types'

function sanitize(html: string): string {
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      'p',
      'br',
      'strong',
      'em',
      'code',
      'a',
      'ul',
      'ol',
      'li',
      'h1',
      'h2',
      'input',
      'label',
      'div',
    ],
    ALLOWED_ATTR: [
      'href',
      'rel',
      'target',
      'type',
      'checked',
      'disabled',
      'data-checked',
      'data-type',
      'class',
    ],
  })
}

const props = defineProps<{
  updates: PublicIncidentUpdate[]
}>()

// Updates already arrive newest-first from the backend; render in that order.
const ordered = computed(() => [...props.updates])

const STATUS_LABEL: Record<PublicIncidentUpdate['status'], string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}

const STATUS_DOT: Record<PublicIncidentUpdate['status'], string> = {
  investigating: 'bg-orange-500',
  identified: 'bg-amber-500',
  monitoring: 'bg-blue-500',
  resolved: 'bg-emerald-500',
}

function fmtPostedAt(iso: string): string {
  try {
    const d = new Date(iso)
    const datePart = d.toLocaleString('en-US', {
      month: 'short',
      day: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'UTC',
    })
    return `${datePart} UTC`
  } catch {
    return iso
  }
}

function relativeAgo(iso: string): string {
  try {
    const ms = Date.now() - new Date(iso).getTime()
    if (ms < 60_000) return 'just now'
    if (ms < 3_600_000) {
      const m = Math.round(ms / 60_000)
      return `${m} minute${m === 1 ? '' : 's'} ago`
    }
    if (ms < 86_400_000) {
      const h = Math.round(ms / 3_600_000)
      return `${h} hour${h === 1 ? '' : 's'} ago`
    }
    const d = Math.round(ms / 86_400_000)
    return `${d} day${d === 1 ? '' : 's'} ago`
  } catch {
    return ''
  }
}
</script>

<template>
  <ol class="space-y-8" data-testid="incident-timeline">
    <li
      v-for="u in ordered"
      :key="u.id"
      class="grid grid-cols-1 md:grid-cols-[160px_1fr] gap-3 md:gap-8"
      :data-update-status="u.status"
    >
      <div class="flex items-center gap-2 md:items-start md:pt-0.5">
        <span :class="['size-2 rounded-full', STATUS_DOT[u.status]]" />
        <h3 class="text-base font-semibold text-gray-900">{{ STATUS_LABEL[u.status] }}</h3>
      </div>
      <div>
        <div
          class="text-base text-gray-900 leading-relaxed prose prose-base max-w-none"
          v-html="sanitize(u.message)"
        />
        <p class="mt-1 text-sm text-gray-500">
          Posted {{ relativeAgo(u.posted_at) }}. {{ fmtPostedAt(u.posted_at) }}
        </p>
      </div>
    </li>
  </ol>
</template>

<style>
ul.rt-task-list {
  list-style: none;
  padding-left: 0;
}
li.rt-task-item {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin: 0.25rem 0;
}
li.rt-task-item > label {
  margin-top: 0.2rem;
  flex-shrink: 0;
  pointer-events: none;
}
li.rt-task-item > label > input[type='checkbox'] {
  width: 1rem;
  height: 1rem;
  accent-color: #4f46e5;
  pointer-events: none;
}
li.rt-task-item > div > p {
  margin: 0;
}
li.rt-task-item[data-checked='true'] > div {
  color: #94a3b8;
  text-decoration: line-through;
}
</style>

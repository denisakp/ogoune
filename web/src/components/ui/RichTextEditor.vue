<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import { ref, watch, onBeforeUnmount, computed } from 'vue'
import DOMPurify from 'dompurify'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  minHeight?: string
}>()

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const preview = ref(false)

const editor = useEditor({
  content: props.modelValue || '',
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2] },
      codeBlock: false,
      blockquote: false,
    }),
    Link.configure({
      openOnClick: false,
      autolink: true,
      HTMLAttributes: { rel: 'noopener noreferrer', target: '_blank' },
    }),
    TaskList.configure({ HTMLAttributes: { class: 'rt-task-list' } }),
    TaskItem.configure({ nested: true, HTMLAttributes: { class: 'rt-task-item' } }),
  ],
  editorProps: {
    attributes: {
      class:
        'prose prose-sm max-w-none focus:outline-none px-3 py-2 text-slate-900',
    },
  },
  onUpdate({ editor }) {
    const html = editor.isEmpty ? '' : editor.getHTML()
    emit('update:modelValue', html)
  },
})

watch(
  () => props.modelValue,
  (val) => {
    const current = editor.value?.getHTML() ?? ''
    if (editor.value && val !== current) {
      editor.value.commands.setContent(val || '', { emitUpdate: false })
    }
  },
)

onBeforeUnmount(() => {
  editor.value?.destroy()
})

function isActive(name: string, attrs?: Record<string, unknown>) {
  return editor.value?.isActive(name, attrs) ?? false
}

function toggleBold() { editor.value?.chain().focus().toggleBold().run() }
function toggleItalic() { editor.value?.chain().focus().toggleItalic().run() }
function toggleCode() { editor.value?.chain().focus().toggleCode().run() }
function toggleBullet() { editor.value?.chain().focus().toggleBulletList().run() }
function toggleOrdered() { editor.value?.chain().focus().toggleOrderedList().run() }
function toggleTaskList() { editor.value?.chain().focus().toggleTaskList().run() }
function toggleH1() { editor.value?.chain().focus().toggleHeading({ level: 1 }).run() }
function toggleH2() { editor.value?.chain().focus().toggleHeading({ level: 2 }).run() }

function setLink() {
  if (!editor.value) return
  const prev = editor.value.getAttributes('link').href
  const url = window.prompt('URL', prev || 'https://')
  if (url === null) return
  if (url === '') {
    editor.value.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }
  editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}

function togglePreview() { preview.value = !preview.value }

const sanitizedHtml = computed(() =>
  DOMPurify.sanitize(props.modelValue || '', {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'a', 'ul', 'ol', 'li', 'h1', 'h2', 'input', 'label', 'div'],
    ALLOWED_ATTR: ['href', 'rel', 'target', 'type', 'checked', 'disabled', 'data-checked', 'data-type', 'class'],
  }),
)

interface TbItem {
  key: string
  title: string
  testid: string
  isText?: string
  iconPath?: string
  active?: () => boolean
  onClick: () => void
}

const items: TbItem[] = [
  { key: 'bold', title: 'Bold', testid: 'rt-bold', isText: 'B', active: () => isActive('bold'), onClick: toggleBold },
  { key: 'italic', title: 'Italic', testid: 'rt-italic', isText: 'I', active: () => isActive('italic'), onClick: toggleItalic },
  { key: 'h1', title: 'Heading 1', testid: 'rt-h1', isText: 'H₁', active: () => isActive('heading', { level: 1 }), onClick: toggleH1 },
  { key: 'h2', title: 'Heading 2', testid: 'rt-h2', isText: 'H₂', active: () => isActive('heading', { level: 2 }), onClick: toggleH2 },
]
</script>

<template>
  <div
    class="rounded-md border border-slate-300 bg-white overflow-hidden"
    data-testid="rich-text-editor"
  >
    <div
      class="flex items-center gap-0.5 border-b border-slate-200 bg-white px-2 py-1.5"
      role="toolbar"
    >
      <!-- Text-style buttons -->
      <button
        v-for="it in items"
        :key="it.key"
        type="button"
        class="h-8 min-w-8 px-2 inline-flex items-center justify-center rounded-md text-sm text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': it.active && it.active() }"
        :title="it.title"
        :data-testid="it.testid"
        @click="it.onClick"
      >
        <span class="font-semibold leading-none">{{ it.isText }}</span>
      </button>

      <!-- Bullet list -->
      <button
        type="button"
        class="h-8 w-8 inline-flex items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': isActive('bulletList') }"
        title="Bullet list"
        data-testid="rt-bullet"
        @click="toggleBullet"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="8" y1="6" x2="21" y2="6" />
          <line x1="8" y1="12" x2="21" y2="12" />
          <line x1="8" y1="18" x2="21" y2="18" />
          <circle cx="3.5" cy="6" r="1" fill="currentColor" stroke="none" />
          <circle cx="3.5" cy="12" r="1" fill="currentColor" stroke="none" />
          <circle cx="3.5" cy="18" r="1" fill="currentColor" stroke="none" />
        </svg>
      </button>

      <!-- Numbered list -->
      <button
        type="button"
        class="h-8 w-8 inline-flex items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': isActive('orderedList') }"
        title="Numbered list"
        data-testid="rt-ordered"
        @click="toggleOrdered"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="10" y1="6" x2="21" y2="6" />
          <line x1="10" y1="12" x2="21" y2="12" />
          <line x1="10" y1="18" x2="21" y2="18" />
          <path d="M4 6h1v4" />
          <path d="M4 10h2" />
          <path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1" />
        </svg>
      </button>

      <!-- Task list (checkboxes) -->
      <button
        type="button"
        class="h-8 w-8 inline-flex items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': isActive('taskList') }"
        title="Task list"
        data-testid="rt-tasks"
        @click="toggleTaskList"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 7 5 9 9 5" />
          <line x1="13" y1="6" x2="21" y2="6" />
          <polyline points="3 14 5 16 9 12" />
          <line x1="13" y1="13" x2="21" y2="13" />
          <polyline points="3 21 5 23 9 19" />
          <line x1="13" y1="20" x2="21" y2="20" />
        </svg>
      </button>

      <!-- Link -->
      <button
        type="button"
        class="h-8 w-8 inline-flex items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': isActive('link') }"
        title="Link"
        data-testid="rt-link"
        @click="setLink"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10 13a5 5 0 0 0 7.07 0l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
          <path d="M14 11a5 5 0 0 0-7.07 0l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
        </svg>
      </button>

      <!-- Inline code -->
      <button
        type="button"
        class="h-8 w-8 inline-flex items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': isActive('code') }"
        title="Inline code"
        data-testid="rt-code"
        @click="toggleCode"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="16 18 22 12 16 6" />
          <polyline points="8 6 2 12 8 18" />
        </svg>
      </button>

      <!-- Preview toggle -->
      <button
        type="button"
        class="ml-auto h-8 inline-flex items-center gap-1.5 px-2.5 rounded-md text-sm text-slate-600 hover:bg-slate-100 hover:text-slate-900"
        :class="{ 'bg-slate-100 text-slate-900': preview }"
        :title="preview ? 'Back to edit' : 'Preview'"
        data-testid="rt-preview"
        @click="togglePreview"
      >
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
        {{ preview ? 'Edit' : 'Preview' }}
      </button>
    </div>

    <div v-show="!preview" :style="{ minHeight: minHeight || '120px' }">
      <EditorContent :editor="editor" />
      <p
        v-if="editor && editor.isEmpty && placeholder"
        class="px-3 -mt-9 pointer-events-none text-sm text-slate-400"
        aria-hidden="true"
      >
        {{ placeholder }}
      </p>
    </div>

    <div
      v-if="preview"
      class="prose prose-sm max-w-none px-3 py-3"
      :style="{ minHeight: minHeight || '120px' }"
      v-html="sanitizedHtml"
    />
  </div>
</template>

<style>
.ProseMirror {
  min-height: inherit;
  outline: none;
}
.ProseMirror p {
  margin: 0 0 0.5rem 0;
}
.ProseMirror p:last-child {
  margin-bottom: 0;
}
.ProseMirror h1 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0.5rem 0;
}
.ProseMirror h2 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0.5rem 0;
}
.ProseMirror ul {
  list-style: disc;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}
.ProseMirror ol {
  list-style: decimal;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}
.ProseMirror code {
  background: #f1f5f9;
  padding: 0 0.25rem;
  border-radius: 3px;
  font-size: 0.875em;
}
.ProseMirror a {
  color: #4f46e5;
  text-decoration: underline;
}
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
}
li.rt-task-item > label > input[type="checkbox"] {
  width: 1rem;
  height: 1rem;
  accent-color: #4f46e5;
  cursor: pointer;
}
li.rt-task-item > div {
  flex: 1;
  min-width: 0;
}
li.rt-task-item > div > p {
  margin: 0;
}
li.rt-task-item[data-checked="true"] > div {
  color: #94a3b8;
  text-decoration: line-through;
}
</style>

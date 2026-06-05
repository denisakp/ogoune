<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import { watch, onBeforeUnmount } from 'vue'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  minHeight?: string
}>()

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const editor = useEditor({
  content: props.modelValue || '',
  extensions: [
    StarterKit.configure({
      heading: { levels: [2, 3] },
      codeBlock: false,
    }),
    Link.configure({
      openOnClick: false,
      autolink: true,
      HTMLAttributes: { rel: 'noopener noreferrer', target: '_blank' },
    }),
  ],
  editorProps: {
    attributes: {
      class:
        'prose prose-sm max-w-none focus:outline-none px-3 py-2 min-h-[120px] text-slate-900',
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

function toggleBold() { editor.value?.chain().focus().toggleBold().run() }
function toggleItalic() { editor.value?.chain().focus().toggleItalic().run() }
function toggleCode() { editor.value?.chain().focus().toggleCode().run() }
function toggleBullet() { editor.value?.chain().focus().toggleBulletList().run() }
function toggleOrdered() { editor.value?.chain().focus().toggleOrderedList().run() }
function toggleH2() { editor.value?.chain().focus().toggleHeading({ level: 2 }).run() }
function toggleH3() { editor.value?.chain().focus().toggleHeading({ level: 3 }).run() }
function toggleBlockquote() { editor.value?.chain().focus().toggleBlockquote().run() }

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

function isActive(name: string, attrs?: Record<string, unknown>) {
  return editor.value?.isActive(name, attrs) ?? false
}
</script>

<template>
  <div
    class="rounded-md border border-slate-300 bg-white overflow-hidden"
    data-testid="rich-text-editor"
  >
    <div
      class="flex flex-wrap items-center gap-1 border-b border-slate-200 bg-slate-50 px-2 py-1"
      role="toolbar"
    >
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-sm font-bold"
        :class="{ 'bg-slate-200': isActive('bold') }"
        title="Bold (Ctrl+B)"
        data-testid="rt-bold"
        @click="toggleBold"
      >B</button>
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-sm italic"
        :class="{ 'bg-slate-200': isActive('italic') }"
        title="Italic (Ctrl+I)"
        data-testid="rt-italic"
        @click="toggleItalic"
      >I</button>
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-xs font-mono"
        :class="{ 'bg-slate-200': isActive('code') }"
        title="Inline code"
        data-testid="rt-code"
        @click="toggleCode"
      >‹/›</button>
      <span class="mx-1 h-5 w-px bg-slate-300" />
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-xs font-semibold"
        :class="{ 'bg-slate-200': isActive('heading', { level: 2 }) }"
        title="Heading 2"
        data-testid="rt-h2"
        @click="toggleH2"
      >H2</button>
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-xs font-semibold"
        :class="{ 'bg-slate-200': isActive('heading', { level: 3 }) }"
        title="Heading 3"
        data-testid="rt-h3"
        @click="toggleH3"
      >H3</button>
      <span class="mx-1 h-5 w-px bg-slate-300" />
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-sm"
        :class="{ 'bg-slate-200': isActive('bulletList') }"
        title="Bullet list"
        data-testid="rt-bullet"
        @click="toggleBullet"
      >•</button>
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-xs font-semibold"
        :class="{ 'bg-slate-200': isActive('orderedList') }"
        title="Numbered list"
        data-testid="rt-ordered"
        @click="toggleOrdered"
      >1.</button>
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-sm"
        :class="{ 'bg-slate-200': isActive('blockquote') }"
        title="Quote"
        data-testid="rt-quote"
        @click="toggleBlockquote"
      >❝</button>
      <span class="mx-1 h-5 w-px bg-slate-300" />
      <button
        type="button"
        class="size-7 inline-flex items-center justify-center rounded hover:bg-slate-200 text-sm"
        :class="{ 'bg-slate-200': isActive('link') }"
        title="Link"
        data-testid="rt-link"
        @click="setLink"
      >🔗</button>
    </div>

    <EditorContent :editor="editor" :style="{ minHeight: minHeight || '120px' }" />

    <p
      v-if="editor && editor.isEmpty && placeholder"
      class="px-3 -mt-9 pointer-events-none text-sm text-slate-400"
      aria-hidden="true"
    >
      {{ placeholder }}
    </p>
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
.ProseMirror h2 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0.5rem 0;
}
.ProseMirror h3 {
  font-size: 1rem;
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
.ProseMirror blockquote {
  border-left: 3px solid #cbd5e1;
  padding-left: 0.75rem;
  color: #475569;
  margin: 0.5rem 0;
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
</style>

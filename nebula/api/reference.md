---
aside: false
outline: false
title: API Reference
---

<script setup>
import { useOpenapi } from 'vitepress-openapi/client'
import spec from '../../api/openapi/v1.json'

useOpenapi({ spec })
</script>

<OASpec />

import type { EnhanceAppContext } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import { theme } from 'vitepress-openapi/client'
import 'vitepress-openapi/dist/style.css'
import './custom.css'

// Extend the default VitePress theme and register the vitepress-openapi
// components (OASpec, OAOperation, …) globally so markdown pages can render
// the OpenAPI reference from api/openapi/v1.json.
export default {
  extends: DefaultTheme,
  async enhanceApp(ctx: EnhanceAppContext) {
    theme.enhanceApp(ctx)
  },
}

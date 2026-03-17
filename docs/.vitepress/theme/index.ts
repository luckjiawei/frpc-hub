import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import { h } from 'vue'
import ScreenshotGallery from './components/ScreenshotGallery.vue'
import AsideAd from './components/AsideAd.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  Layout() {
    return h(DefaultTheme.Layout, null, {
      'aside-outline-after': () => h(AsideAd),
    })
  },
  enhanceApp({ app }) {
    app.component('ScreenshotGallery', ScreenshotGallery)
  },
} satisfies Theme

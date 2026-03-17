<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Image {
  src: string
  caption: string
}

const images: Image[] = [
  { src: '/22.png', caption: '服务器管理 — 实时延迟与地理位置' },
  { src: '/33.png', caption: '代理管理 — 可视化配置所有隧道' },
  { src: '/44.png', caption: '代理详情 — 实时连接日志' },
  { src: '/55.png', caption: '一键导入 — 支持 TOML 配置文件' },
  { src: '/66.png', caption: '添加代理 — 表单填写界面' },
]

const activeIndex = ref<number | null>(null)
const active = computed(() => activeIndex.value !== null ? images[activeIndex.value] : null)

function open(i: number) {
  activeIndex.value = i
  document.body.style.overflow = 'hidden'
}

function close() {
  activeIndex.value = null
  document.body.style.overflow = ''
}

function prev() {
  if (activeIndex.value !== null && activeIndex.value > 0) activeIndex.value--
}

function next() {
  if (activeIndex.value !== null && activeIndex.value < images.length - 1) activeIndex.value++
}

function onKey(e: KeyboardEvent) {
  if (activeIndex.value === null) return
  if (e.key === 'Escape') close()
  if (e.key === 'ArrowLeft') prev()
  if (e.key === 'ArrowRight') next()
}

onMounted(() => document.addEventListener('keydown', onKey))
onUnmounted(() => {
  document.removeEventListener('keydown', onKey)
  document.body.style.overflow = ''
})
</script>

<template>
  <div class="sg-gallery">
    <div
      v-for="(img, i) in images"
      :key="img.src"
      class="sg-item"
      @click="open(i)"
    >
      <img :src="img.src" :alt="img.caption" loading="lazy" />
      <div class="sg-caption">{{ img.caption }}</div>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="sg-fade">
      <div v-if="active" class="sg-lightbox" @click="close">
        <button class="sg-close" aria-label="关闭" @click.stop="close">✕</button>
        <div class="sg-lightbox-inner" @click.stop>
          <img :src="active.src" :alt="active.caption" />
          <p class="sg-lightbox-caption">{{ active.caption }}</p>
        </div>
        <button
          v-if="activeIndex! > 0"
          class="sg-nav sg-nav-prev"
          aria-label="上一张"
          @click.stop="prev"
        >‹</button>
        <button
          v-if="activeIndex! < images.length - 1"
          class="sg-nav sg-nav-next"
          aria-label="下一张"
          @click.stop="next"
        >›</button>
      </div>
    </Transition>
  </Teleport>
</template>

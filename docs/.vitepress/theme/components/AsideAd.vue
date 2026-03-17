<template>
  <div class="aside-ad">
    <a href="https://www.rainyun.com/s=ljw_?podux" target="_blank" rel="noopener noreferrer">
      <div class="aside-ad-inner">
        <div class="aside-ad-badge">赞助商</div>
        <div class="aside-ad-content">
          <div class="aside-ad-title">☁️ 雨云</div>
          <div class="aside-ad-desc">价格实惠的国内云服务器，轻松部署 frps 服务</div>
          <div class="aside-ad-cta">立即查看 →</div>
        </div>
      </div>
    </a>
  </div>
</template>

<style scoped>
.aside-ad {
  margin-top: 24px;
}

.aside-ad a {
  text-decoration: none;
}

.aside-ad-inner {
  position: relative;
  border: 1px solid var(--vp-c-brand-1);
  border-radius: 8px;
  padding: 12px 14px;
  background: var(--vp-c-bg-soft);
  transition: background 0.2s, border-color 0.2s;
}

.aside-ad-inner:hover {
  background: var(--vp-c-brand-soft);
  border-color: var(--vp-c-brand-2);
}

.aside-ad-badge {
  font-size: 10px;
  color: var(--vp-c-text-3);
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.aside-ad-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  margin-bottom: 4px;
}

.aside-ad-desc {
  font-size: 12px;
  color: var(--vp-c-text-2);
  line-height: 1.5;
  margin-bottom: 8px;
}

.aside-ad-cta {
  font-size: 12px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
}
</style>

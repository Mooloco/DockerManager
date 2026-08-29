<template>
  <div class="sparkline">
    <svg :viewBox="`0 0 ${width} ${height}`" preserveAspectRatio="none" class="svg">
      <!-- 网格线 -->
      <line v-for="y in 3" :key="y" :x1="0" :x2="width" :y1="(y * height) / 4" :y2="(y * height) / 4" class="grid" />
      <polyline v-if="points.length > 1" :points="polyPoints" class="line" :style="{ stroke: color }" />
      <circle v-if="points.length" :cx="lastX" :cy="lastY" r="3" :fill="color" />
    </svg>
    <div class="axis">
      <span>{{ format(max) }} {{ unit }}</span>
      <span>0</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  points: number[]
  max: number
  unit?: string
  color?: string
}>()

const width = 400
const height = 90

const normMax = computed(() => (props.max > 0 ? props.max : 1))

const polyPoints = computed(() => {
  const pts = props.points.slice(-60)
  if (pts.length < 2) return ''
  const step = width / (pts.length - 1)
  return pts
    .map((v, i) => {
      const x = i * step
      const y = height - (Math.min(v, normMax.value) / normMax.value) * (height - 8) - 4
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})

const lastX = computed(() => {
  const pts = props.points.slice(-60)
  return pts.length > 1 ? width : 0
})

const lastY = computed(() => {
  const pts = props.points.slice(-60)
  if (!pts.length) return 0
  const v = pts[pts.length - 1]
  return height - (Math.min(v, normMax.value) / normMax.value) * (height - 8) - 4
})

function format(v: number): string {
  if (v >= 1000) return `${(v / 1000).toFixed(1)}k`
  return String(Math.round(v))
}
</script>

<style scoped>
.sparkline {
  width: 100%;
}

.svg {
  width: 100%;
  height: 90px;
  display: block;
}

.grid {
  stroke: var(--el-border-color-lighter);
  stroke-width: 0.5;
  stroke-dasharray: 3 3;
}

.line {
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.axis {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}
</style>

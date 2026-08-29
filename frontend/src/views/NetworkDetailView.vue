<template>
  <div class="page-container" v-loading="!loaded">
    <div class="detail-head">
      <el-button text :icon="ArrowLeft" @click="router.push('/networks')">返回</el-button>
      <span class="name">{{ detail?.Name || '网络详情' }}</span>
      <el-tag v-if="detail" size="small" :type="detail.Driver === 'bridge' ? 'primary' : 'info'" effect="plain">
        {{ detail.Driver }}
      </el-tag>
    </div>

    <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false" class="mb-16" />

    <template v-if="detail">
      <!-- 基本信息 -->
      <el-card shadow="never" class="mb-16">
        <template #header>
          <div class="card-header"><el-icon><InfoFilled /></el-icon><span>基本信息</span></div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ detail.Name }}</el-descriptions-item>
          <el-descriptions-item label="作用域">{{ detail.Scope }}</el-descriptions-item>
          <el-descriptions-item label="ID" :span="2"><span class="mono">{{ detail.Id }}</span></el-descriptions-item>
          <el-descriptions-item label="Internal">
            <el-tag :type="detail.Internal ? 'warning' : 'info'" size="small">{{ detail.Internal ? '是' : '否' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="IPv6">
            <el-tag :type="detail.EnableIPv6 ? 'success' : 'info'" size="small">{{ detail.EnableIPv6 ? '是' : '否' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Attachable">{{ detail.Attachable ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="Ingress">{{ detail.Ingress ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="子网">{{ subnet }}</el-descriptions-item>
          <el-descriptions-item label="网关">{{ gateway }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 连接容器 -->
      <el-card shadow="never">
        <template #header>
          <div class="card-header"><el-icon><Box /></el-icon><span>连接容器({{ containerList.length }})</span></div>
        </template>
        <el-table v-if="containerList.length" :data="containerList" size="small" stripe>
          <el-table-column label="容器" min-width="160">
            <template #default="{ row }">
              <span class="mono">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column label="IPv4 地址" width="180">
            <template #default="{ row }">
              <span class="mono dim">{{ row.ipv4 || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="IPv6 地址" min-width="200">
            <template #default="{ row }">
              <span class="mono dim">{{ row.ipv6 || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="MAC 地址" min-width="180">
            <template #default="{ row }">
              <span class="mono dim">{{ row.mac || '-' }}</span>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="该网络当前没有容器连接" :image-size="60" />
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, InfoFilled, Box } from '@element-plus/icons-vue'
import { networkApi } from '../api'

interface NetEndpoint {
  name: string
  ipv4: string
  ipv6: string
  mac: string
}

interface NetInspect {
  Name: string
  Id: string
  Scope: string
  Driver: string
  EnableIPv6: boolean
  Internal: boolean
  Attachable: boolean
  Ingress: boolean
  IPAM?: { Config?: { Subnet?: string; Gateway?: string }[] }
  Containers?: Record<string, { Name?: string; IPv4Address?: string; IPv6Address?: string; MacAddress?: string }>
}

const route = useRoute()
const router = useRouter()
const detail = ref<NetInspect | null>(null)
const loaded = ref(false)
const loadError = ref('')

const subnet = computed(() => detail.value?.IPAM?.Config?.[0]?.Subnet || '-')
const gateway = computed(() => detail.value?.IPAM?.Config?.[0]?.Gateway || '-')

const containerList = computed<NetEndpoint[]>(() => {
  const out: NetEndpoint[] = []
  for (const c of Object.values(detail.value?.Containers || {})) {
    out.push({
      name: c?.Name || '',
      ipv4: c?.IPv4Address || '',
      ipv6: c?.IPv6Address || '',
      mac: c?.MacAddress || '',
    })
  }
  return out
})

onMounted(async () => {
  try {
    detail.value = (await networkApi.inspect(route.params.id as string)) as unknown as NetInspect
  } catch (e: any) {
    loadError.value = e.message || '获取网络详情失败'
    ElMessage.error(loadError.value)
  } finally {
    loaded.value = true
  }
})
</script>

<style scoped>
.mb-16 {
  margin-bottom: 16px;
}

.detail-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.detail-head .name {
  font-size: 17px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.dim {
  color: var(--el-text-color-secondary);
}
</style>

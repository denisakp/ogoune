<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import FeedbackModal from '@/components/FeedbackModal.vue'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  DashboardOutlined,
  AlertOutlined,
  SettingOutlined,
  ApiOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const isSidebarOpen = ref(false)

const navigation = [
  { name: 'Monitoring', path: '/monitors', key: '1', icon: DashboardOutlined },
  { name: 'Incidents', path: '/incidents', key: '2', icon: AlertOutlined },
  { name: 'Status Page', path: '/status', key: '3', icon: ApiOutlined },
  { name: 'Maintenance', path: '/maintenance', key: '4', icon: ToolOutlined },
  { name: 'Settings', path: '/settings', key: '5', icon: SettingOutlined },
]

// Check if current route requires layout
const requiresLayout = computed(() => {
  return route.meta.requiresLayout !== false
})

const menuItems = computed(() =>
  navigation.map((item) => ({
    key: item.key,
    label: item.name,
    icon: () => h(item.icon),
  })),
)

const isDektop = computed(() => window.innerWidth >= 1024)

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const handleMenuClick = (key: string) => {
  const item = navigation[parseInt(key) - 1]
  if (item) {
    if (item.path === '/status') {
      // Open status page in new tab
      window.open(item.path, '_blank')
    } else {
      // Navigate normally
      router.push(item.path)
    }
  }
}
</script>

<template>
  <!-- Public routes (no layout) -->
  <div v-if="!requiresLayout">
    <FeedbackModal />
    <router-view />
  </div>

  <!-- Admin routes (with layout) -->
  <a-layout v-else style="min-height: 100vh">
    <FeedbackModal />
    <!-- Mobile Drawer Sidebar -->
    <a-drawer
      v-model:open="isSidebarOpen"
      title="Navigation"
      placement="left"
      :closable="true"
      :mask="true"
      class="lg:hidden"
    >
      <a-menu
        :items="menuItems"
        mode="vertical"
        theme="dark"
        @click="
          (e: any) => {
            handleMenuClick(e.key)
            isSidebarOpen = false
          }
        "
      />

      <!-- Logout button in drawer -->
      <div
        style="
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 16px;
          border-top: 1px solid #f0f0f0;
        "
      >
        <a-button type="text" danger block @click="handleLogout">
          <template #icon>
            <LogoutOutlined />
          </template>
          Logout
        </a-button>
      </div>
    </a-drawer>

    <!-- Desktop Sidebar -->
    <a-layout-sider
      class="hidden lg:block sidebar-container"
      :width="256"
      style="background: #fafafa; position: fixed; left: 0; top: 0; bottom: 0; z-index: 100"
    >
      <!-- Logo at top (sans bordure) -->
      <div style="padding: 24px 16px">
        <div style="font-size: 20px; font-weight: bold; color: #1890ff">Pulseguard</div>
      </div>

      <!-- Menu -->
      <a-menu
        :items="menuItems"
        mode="vertical"
        style="border: none; background: transparent"
        @click="
          (e: any) => {
            handleMenuClick(e.key)
          }
        "
      />

      <!-- Logout button at bottom -->
      <div
        style="
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 15px;
          border-top: 1px solid #e8e8e8;
          background: #fafafa;
        "
      >
        <a-button type="text" danger block @click="handleLogout">
          <template #icon><LogoutOutlined /></template>
          Logout
        </a-button>
      </div>
    </a-layout-sider>

    <!-- Main Content avec margin-left pour le desktop -->
    <a-layout style="background: #fff" :style="{ marginLeft: isDektop ? '256px' : '0' }">
      <!-- Mobile menu toggle -->
      <div class="lg:hidden" style="padding: 16px; background: #fff">
        <button
          @click="isSidebarOpen = !isSidebarOpen"
          style="
            border: none;
            background: none;
            cursor: pointer;
            font-size: 18px;
            display: flex;
            align-items: center;
            gap: 8px;
          "
        >
          <MenuUnfoldOutlined v-if="!isSidebarOpen" />
          <MenuFoldOutlined v-else />
          <span style="font-weight: 600">Menu</span>
        </button>
      </div>

      <a-layout-content style="padding: 24px; background: #fff; min-height: calc(100vh - 70px)">
        <router-view />
      </a-layout-content>

      <!-- Footer fixe en bas -->
      <a-layout-footer style="text-align: center; background: #fff; padding: 16px">
        Pulse Guard ©{{ new Date().getUTCFullYear() }} By
        <a target="_blank" href="https://github.com/denisakp">Denis Yaovi</a>
        ·
        <a
          target="_blank"
          href="https://kawa-bunga.notion.site/2d1e5ad0a17d80dc8859e77817d901e3"
          rel="noopener"
          >Share Feedback</a
        >
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.sidebar-container {
  height: 100vh;
  overflow-y: auto;
}

@media (max-width: 1024px) {
  .hidden.lg\\:block {
    display: none !important;
  }

  .lg\\:hidden {
    display: block !important;
  }
}

@media (min-width: 1024px) {
  .hidden.lg\\:block {
    display: block !important;
  }

  .lg\\:hidden {
    display: none !important;
  }
}
</style>

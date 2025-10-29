<script setup lang="ts">
import { ref } from 'vue'
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons-vue'

const isSidebarOpen = ref(false)

const navigation = [
  { name: 'Monitoring', path: '/monitors', key: '1', icon: '📊' },
  { name: 'Incidents', path: '/incidents', key: '2', icon: '🏷️' },
  { name: 'Settings', path: '/Settings', key: '3', icon: '🔌' },
  { name: 'Activities', path: '/activities', key: '4', icon: '📜' },
]

const menuItems = navigation.map((item) => ({
  key: item.key,
  label: item.name,
}))
</script>

<template>
  <a-layout style="min-height: 100vh">
    <!-- Header / Navbar -->
    <a-layout-header
      class="header"
      style="
        background: #fff;
        padding: 0 24px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      "
    >
      <div style="display: flex; align-items: center; gap: 16px; flex: 1">
        <button
          @click="isSidebarOpen = !isSidebarOpen"
          style="display: none; border: none; background: none; cursor: pointer; font-size: 18px"
          class="md:hidden"
        >
          <MenuUnfoldOutlined v-if="isSidebarOpen" />
          <MenuFoldOutlined v-else />
        </button>
        <div style="font-size: 20px; font-weight: bold; color: #1890ff">🔍 Pulseguard</div>
      </div>
      <a-tag color="blue">Ant Design Vue</a-tag>
    </a-layout-header>

    <a-layout>
      <!-- Sidebar -->
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
          @click="
            (e) => {
              const item = navigation[parseInt(e.key) - 1]
              $router.push(item.path)
              isSidebarOpen = false
            }
          "
        />
      </a-drawer>

      <a-layout-sider class="hidden lg:block" :width="256" style="background: #fafafa">
        <a-menu
          :items="menuItems"
          mode="vertical"
          @click="
            (e) => {
              const item = navigation[parseInt(e.key) - 1]
              $router.push(item.path)
            }
          "
        />
      </a-layout-sider>

      <!-- Main Content -->
      <a-layout>
        <a-layout-content style="padding: 24px">
          <router-view />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

@media (max-width: 768px) {
  :deep(.ant-layout-sider) {
    display: none;
  }
}
</style>

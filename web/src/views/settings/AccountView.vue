<script setup lang="ts">
/**
 * Account view — design fidelity v2.
 * Page-level "Profile" h1 + 3 cards (Profile / Change Password / Danger Zone).
 */
import ProfileSection from '@/components/settings/account/ProfileSection.vue'
import ChangePasswordSection from '@/components/settings/account/ChangePasswordSection.vue'
import DangerZoneSection from '@/components/settings/account/DangerZoneSection.vue'
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-default">Profile</h1>

    <ProfileSection />
    <ChangePasswordSection />
    <DangerZoneSection />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import Sidebar from "./components/Sidebar.vue";
import ErrorBoundary from "./components/ErrorBoundary.vue";
import ToastContainer from "./components/ToastContainer.vue";
import { fetchCatalogs, fetchTheme } from "./lib/api";
import type { Catalog } from "./lib/types";

const catalogs = ref<Catalog[]>([]);
const loading = ref(true);
const sidebarOpen = ref(false);
const isMobile = ref(false);

function checkMobile() {
  isMobile.value = window.innerWidth < 768;
  if (!isMobile.value) {
    sidebarOpen.value = false;
  }
}

onMounted(() => {
  checkMobile();
  window.addEventListener("resize", checkMobile);
});
onUnmounted(() => {
  window.removeEventListener("resize", checkMobile);
});

// Theme token name → CSS custom property mapping.
const themeMap: Record<string, string> = {
  background: "--dops-background",
  backgroundPanel: "--dops-backgroundPanel",
  backgroundElement: "--dops-backgroundElement",
  backgroundHover: "--dops-backgroundHover",
  text: "--dops-text",
  textMuted: "--dops-textMuted",
  textSubtle: "--dops-textSubtle",
  primary: "--dops-primary",
  primaryMuted: "--dops-primaryMuted",
  border: "--dops-border",
  borderActive: "--dops-borderActive",
  success: "--dops-success",
  successMuted: "--dops-successMuted",
  warning: "--dops-warning",
  warningMuted: "--dops-warningMuted",
  error: "--dops-error",
  errorMuted: "--dops-errorMuted",
};

function applyTheme(colors: Record<string, string>) {
  const root = document.documentElement;
  for (const [token, cssVar] of Object.entries(themeMap)) {
    const value = colors[token];
    if (value && value !== "none") {
      root.style.setProperty(cssVar, value);
    }
  }
}

onMounted(async () => {
  // Load theme and catalogs in parallel.
  const [themeResult] = await Promise.allSettled([
    fetchTheme().then((t) => applyTheme(t.colors)),
    fetchCatalogs()
      .then((c) => (catalogs.value = c))
      .finally(() => (loading.value = false)),
  ]);
  if (themeResult.status === "rejected") {
    console.error("Failed to load theme:", themeResult.reason);
  }
});

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && sidebarOpen.value) {
    sidebarOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener("keydown", onKeydown);
});
onUnmounted(() => {
  document.removeEventListener("keydown", onKeydown);
});
</script>

<template>
  <div class="flex h-screen">
    <!-- Desktop sidebar (always visible >= 768px) -->
    <div class="hidden md:flex">
      <Sidebar
        :catalogs="catalogs"
        :loading="loading"
        @themeChanged="applyTheme"
      />
    </div>

    <!-- Mobile top bar (visible < 768px) -->
    <div
      v-if="isMobile"
      class="fixed top-0 left-0 right-0 z-40 h-12 bg-bg-panel border-b border-border flex items-center px-4 gap-3"
    >
      <button
        @click="sidebarOpen = !sidebarOpen"
        class="p-1 bg-transparent border-none cursor-pointer text-fg-muted hover:text-fg transition-colors duration-150"
      >
        <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
          <rect x="2" y="4" width="16" height="2" rx="1" />
          <rect x="2" y="9" width="16" height="2" rx="1" />
          <rect x="2" y="14" width="16" height="2" rx="1" />
        </svg>
      </button>
      <span class="text-primary font-mono text-[15px] font-bold tracking-tight">dops</span>
      <span class="text-[15px] font-bold text-fg">runbooks</span>
    </div>

    <!-- Mobile sidebar overlay -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isMobile && sidebarOpen"
        class="fixed inset-0 bg-black/50 z-40"
        @click="sidebarOpen = false"
      />
    </Transition>

    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <div
        v-if="isMobile && sidebarOpen"
        class="fixed top-0 left-0 bottom-0 z-50"
      >
        <Sidebar
          :catalogs="catalogs"
          :loading="loading"
          @themeChanged="applyTheme"
          @update:open="sidebarOpen = $event"
        />
      </div>
    </Transition>

    <!-- Main content -->
    <main class="flex-1 overflow-hidden" :class="isMobile ? 'pt-12' : ''">
      <ErrorBoundary>
        <router-view />
      </ErrorBoundary>
    </main>
    <ToastContainer />
  </div>
</template>

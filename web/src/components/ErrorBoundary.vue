<script setup lang="ts">
import { ref, onErrorCaptured } from "vue";

const error = ref<string>("");
const stack = ref<string>("");

onErrorCaptured((err: Error) => {
  error.value = err.message || "An unexpected error occurred";
  stack.value = err.stack || "";
  console.error("ErrorBoundary caught:", err);
  return false; // prevent propagation
});

function dismiss() {
  error.value = "";
  stack.value = "";
}
</script>

<template>
  <div v-if="error" class="flex items-center justify-center h-full p-8">
    <div class="max-w-[480px] w-full text-center">
      <div class="w-10 h-10 rounded-[10px] bg-error-muted text-error flex items-center justify-center text-xl mx-auto mb-4">
        !
      </div>
      <h2 class="text-base font-bold text-fg mb-2">Something went wrong</h2>
      <p class="text-sm text-fg-muted mb-5 leading-relaxed">{{ error }}</p>
      <div class="flex gap-2.5 justify-center">
        <button
          @click="dismiss"
          class="px-4 py-2 text-[13px] font-semibold rounded-md border border-border text-fg-muted hover:border-fg-subtle hover:text-fg bg-transparent cursor-pointer transition-all duration-150"
        >
          Dismiss
        </button>
        <button
          @click="$router.push('/')"
          class="px-4 py-2 text-[13px] font-semibold rounded-md bg-primary text-black cursor-pointer hover:opacity-90 transition-opacity duration-150 border-none"
        >
          Go to Dashboard
        </button>
      </div>
    </div>
  </div>
  <slot v-else />
</template>

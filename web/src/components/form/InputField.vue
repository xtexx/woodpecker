<template>
  <div class="mt-2 mb-4 flex flex-col">
    <div class="mb-2 flex items-center">
      <label class="text-wp-text-100 font-bold" :for="id" v-bind="$attrs">{{ label }}</label>
      <DocsLink v-if="docsUrl" :topic="label" :url="docsUrl" class="ml-2" />
      <slot v-else-if="$slots.titleActions" name="titleActions" />
    </div>
    <div v-if="$slots.description" class="text-wp-text-alt-100 mb-2 text-sm">
      <slot name="description" />
    </div>
    <slot :id="id" />
  </div>
</template>

<script lang="ts" setup>
import DocsLink from '~/components/atomic/DocsLink.vue';

defineProps<{
  label: string;
  docsUrl?: string;
}>();

const id = (Math.random() + 1).toString(36).substring(7);
</script>

<script lang="ts">
export default {
  inheritAttrs: false,
};
</script>

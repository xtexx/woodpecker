<template>
  <div class="text-wp-text-100 space-y-4">
    <ListItem
      v-for="registry in registries"
      :key="registry.id"
      class="bg-wp-background-200! dark:bg-wp-background-100! items-center"
    >
      <span>{{ registry.address }}</span>
      <IconButton
        :icon="registry.readonly ? 'chevron-right' : 'edit'"
        class="ml-auto h-8 w-8"
        :title="registry.readonly ? $t('registries.view') : $t('registries.edit')"
        @click="editRegistry(registry)"
      />
      <IconButton
        v-if="!registry.readonly"
        icon="trash"
        class="hover:text-wp-error-100 h-8 w-8"
        :is-loading="isDeleting"
        :title="$t('registries.delete')"
        @click="deleteRegistry(registry)"
      />
    </ListItem>

    <div v-if="loading" class="flex justify-center">
      <Icon name="spinner" class="animate-spin" />
    </div>
    <div v-else-if="registries?.length === 0" class="ml-2">{{ $t('registries.none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Registry } from '~/lib/api/types';

const props = defineProps<{
  modelValue: (Registry & { edit?: boolean })[];
  isDeleting: boolean;
  loading: boolean;
}>();

const emit = defineEmits<{
  (event: 'edit', registry: Registry): void;
  (event: 'delete', registry: Registry): void;
}>();

const i18n = useI18n();

const registries = toRef(props, 'modelValue');

function editRegistry(registry: Registry) {
  emit('edit', registry);
}

function deleteRegistry(registry: Registry) {
  // TODO: use proper dialog
  // eslint-disable-next-line no-alert
  if (!confirm(i18n.t('registries.delete_confirm'))) {
    return;
  }
  emit('delete', registry);
}
</script>

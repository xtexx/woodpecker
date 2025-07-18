<template>
  <Settings :title="$t('admin.settings.orgs.orgs')" :description="$t('admin.settings.orgs.desc')">
    <div class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="org in orgs"
        :key="org.id"
        class="bg-wp-background-200! dark:bg-wp-background-100! items-center gap-2"
      >
        <span>{{ org.name }}</span>
        <IconButton
          icon="chevron-right"
          :title="$t('admin.settings.orgs.view')"
          class="ml-auto h-8 w-8"
          :to="{ name: 'org', params: { orgId: org.id } }"
        />
        <IconButton
          icon="settings-outline"
          :title="$t('admin.settings.orgs.org_settings')"
          class="h-8 w-8"
          :to="{ name: 'org-settings', params: { orgId: org.id } }"
        />
        <IconButton
          icon="trash"
          :title="$t('admin.settings.orgs.delete_org')"
          class="hover:text-wp-error-100 ml-2 h-8 w-8"
          :is-loading="isDeleting"
          @click="deleteOrg(org)"
        />
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="orgs?.length === 0" class="ml-2">{{ $t('admin.settings.orgs.none') }}</div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Org } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

async function loadOrgs(page: number): Promise<Org[] | null> {
  return apiClient.getOrgs({ page });
}

const { resetPage, data: orgs, loading } = usePagination(loadOrgs);

const { doSubmit: deleteOrg, isLoading: isDeleting } = useAsyncAction(async (_org: Org) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.orgs.delete_confirm'))) {
    return;
  }

  await apiClient.deleteOrg(_org);
  notifications.notify({ title: t('admin.settings.orgs.deleted'), type: 'success' });
  await resetPage();
});

useWPTitle(computed(() => [t('admin.settings.orgs.orgs'), t('admin.settings.settings')]));
</script>

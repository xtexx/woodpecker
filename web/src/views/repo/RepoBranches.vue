<template>
  <div class="space-y-4">
    <template v-if="branches.length > 0">
      <ListItem
        v-for="branch in branchesWithDefaultBranchFirst"
        :key="branch"
        class="text-wp-text-100"
        :to="{ name: 'repo-branch', params: { branch } }"
      >
        {{ branch }}
        <Badge v-if="branch === repo?.default_branch" :value="$t('default')" class="ml-auto" />
      </ListItem>
    </template>
    <div v-else-if="loading" class="text-wp-text-100 flex justify-center">
      <Icon name="spinner" />
    </div>
    <Panel v-else class="flex justify-center">
      {{ $t('empty_list', { entity: $t('repo.branches') }) }}
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { computed, inject, watch } from 'vue';
import type { Ref } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { usePagination } from '~/compositions/usePaginate';
import type { Repo } from '~/lib/api/types';

const apiClient = useApiClient();

const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

async function loadBranches(page: number): Promise<string[]> {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  return apiClient.getRepoBranches(repo.value.id, { page });
}

const { resetPage, data: branches, loading } = usePagination(loadBranches);

const branchesWithDefaultBranchFirst = computed(() =>
  branches.value.toSorted((a, b) => {
    if (a === repo.value.default_branch) {
      return -1;
    }

    if (b === repo.value.default_branch) {
      return 1;
    }

    return 0;
  }),
);

watch(repo, resetPage);
</script>

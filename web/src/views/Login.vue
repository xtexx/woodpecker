<template>
  <main class="flex h-full w-full flex-col items-center justify-center">
    <Error v-if="errorMessage" class="w-full md:w-3xl">
      <span class="whitespace-pre">{{ errorMessage }}</span>
      <span v-if="errorDescription" class="mt-1 whitespace-pre">{{ errorDescription }}</span>
      <a
        v-if="errorUri"
        :href="errorUri"
        target="_blank"
        class="text-wp-link-100 hover:text-wp-link-200 mt-1 cursor-pointer"
      >
        <span>{{ errorUri }}</span>
      </a>
    </Error>

    <div
      class="min-h-sm border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 flex w-full flex-col overflow-hidden border shadow-sm md:m-8 md:w-3xl md:flex-row md:rounded-md"
    >
      <div class="bg-wp-primary-200 dark:bg-wp-primary-300 flex min-h-48 items-center justify-center md:w-3/5">
        <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="h-32 w-32 md:h-48 md:w-48" />
      </div>
      <div class="flex min-h-48 flex-col items-center justify-center gap-4 p-4 text-center md:w-2/5">
        <h1 class="text-wp-text-100 text-xl">{{ $t('login_to_woodpecker_with') }}</h1>
        <div class="flex flex-col gap-2">
          <Button
            v-for="forge in forges"
            :key="forge.id"
            :start-icon="forge.type === 'addon' ? 'repo' : forge.type"
            class="whitespace-normal!"
            @click="doLogin(forge.id)"
          >
            <div class="mr-2 w-4">
              <img
                v-if="!failedForgeFavicons.has(forge.id)"
                :src="getFaviconUrl(forge)"
                :alt="$t('login_to_woodpecker_with', { forge: getHostFromUrl(forge) })"
                @error="() => failedForgeFavicons.add(forge.id)"
              />
              <Icon v-else :name="forge.type === 'addon' ? 'repo' : forge.type" />
            </div>

            {{ getHostFromUrl(forge) }}
          </Button>
        </div>
      </div>
    </div>
  </main>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import Error from '~/components/atomic/Error.vue';
import Icon from '~/components/atomic/Icon.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Forge } from '~/lib/api/types';

const route = useRoute();
const router = useRouter();
const authentication = useAuthentication();
const i18n = useI18n();
const apiClient = useApiClient();

const forges = ref<Forge[]>([]);

function doLogin(forgeId?: number) {
  const url = typeof route.query.url === 'string' ? route.query.url : '';
  authentication.authenticate(url, forgeId);
}

const authErrorMessages = {
  oauth_error: i18n.t('oauth_error'),
  internal_error: i18n.t('internal_error'),
  registration_closed: i18n.t('registration_closed'),
  access_denied: i18n.t('access_denied'),
  invalid_state: i18n.t('invalid_state'),
  org_access_denied: i18n.t('org_access_denied'),
};

const errorMessage = ref<string>();
const errorDescription = ref<string>(route.query.error_description as string);
const errorUri = ref<string>(route.query.error_uri as string);

onMounted(async () => {
  if (authentication.isAuthenticated) {
    await router.replace({ name: 'home' });
    return;
  }

  forges.value = (await apiClient.getForges()) ?? [];

  if (route.query.error) {
    const error = route.query.error as keyof typeof authErrorMessages;
    errorMessage.value = authErrorMessages[error] ?? error;
  }
});

useWPTitle(computed(() => [i18n.t('login')]));

function getHostFromUrl(forge: Forge) {
  if (!forge.url && !forge.oauth_host) {
    return forge.type.charAt(0).toUpperCase() + forge.type.slice(1);
  }

  const url = new URL(forge.oauth_host ?? forge.url);
  return url.hostname;
}

const failedForgeFavicons = ref(new Set<number>()); // Track which favicons failed to load
function getFaviconUrl(forge: Forge) {
  const url = new URL(forge.oauth_host ?? forge.url);
  return `${url.origin}/favicon.ico`;
}
</script>

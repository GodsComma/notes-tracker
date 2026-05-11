<template>
  <div class="p-8">
    <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 mb-6">
      Search results for "<span class="text-indigo-600 dark:text-indigo-400">{{ query }}</span>"
    </h1>

    <!-- Class filter -->
    <div v-if="classStore.classes.length > 1" class="mb-6">
      <div class="flex flex-wrap gap-2">
        <button
          class="text-xs px-3 py-1.5 rounded-full border transition-colors"
          :class="!filterClassID ? 'bg-indigo-600 text-white border-indigo-600' : 'text-slate-600 dark:text-slate-300 border-slate-300 dark:border-slate-600 hover:border-indigo-400'"
          @click="filterClassID = null"
        >
          All classes
        </button>
        <button
          v-for="cls in classStore.classes"
          :key="cls.id"
          class="text-xs px-3 py-1.5 rounded-full border transition-colors flex items-center gap-1.5"
          :class="filterClassID === cls.id ? 'bg-indigo-600 text-white border-indigo-600' : 'text-slate-600 dark:text-slate-300 border-slate-300 dark:border-slate-600 hover:border-indigo-400'"
          @click="filterClassID = filterClassID === cls.id ? null : cls.id"
        >
          <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: cls.color }" />
          {{ cls.name }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-slate-400 text-sm">Searching…</div>

    <div v-else-if="results.length === 0" class="text-center py-20 text-slate-400 dark:text-slate-500">
      <p class="text-4xl mb-3">🔍</p>
      <p class="text-sm">No results found.</p>
    </div>

    <div v-else class="space-y-3">
      <RouterLink
        v-for="r in results"
        :key="r.note_id"
        :to="`/sources/${r.source_id}`"
        class="block bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-5 hover:border-indigo-300 dark:hover:border-indigo-500 hover:shadow-sm transition-all"
      >
        <div class="flex items-start gap-3 mb-2">
          <span class="w-2.5 h-2.5 rounded-full mt-1.5 shrink-0" :style="{ backgroundColor: r.class_color }" />
          <div class="min-w-0">
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ r.class_name }}</p>
            <p class="font-semibold text-slate-800 dark:text-slate-200 truncate">{{ r.source_title }}</p>
            <p v-if="r.source_date" class="text-xs text-slate-400 dark:text-slate-500">{{ formatDate(r.source_date) }}</p>
          </div>
        </div>
        <!-- eslint-disable-next-line vue/no-v-html -->
        <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed pl-5" v-html="r.snippet" />
      </RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, type SearchResult } from '@/api'
import { useClassStore } from '@/stores/class'

const route = useRoute()
const classStore = useClassStore()

const query = ref(String(route.query.q ?? ''))
const filterClassID = ref<number | null>(null)
const results = ref<SearchResult[]>([])
const loading = ref(false)

async function doSearch() {
  if (!query.value.trim()) return
  loading.value = true
  try {
    results.value = await api.search(query.value, filterClassID.value ?? undefined)
  } finally {
    loading.value = false
  }
}

watch(() => route.query.q, q => {
  query.value = String(q ?? '')
  doSearch()
}, { immediate: true })

watch(filterClassID, () => doSearch())

function formatDate(d: string) {
  if (!d) return ''
  return new Date(d + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>

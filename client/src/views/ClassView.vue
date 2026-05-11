<template>
  <div class="p-8">
    <!-- Header -->
    <div v-if="cls" class="mb-6">
      <div class="flex items-center gap-3 mb-1">
        <span class="w-4 h-4 rounded-full shrink-0" :style="{ backgroundColor: cls.color }" />
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ cls.name }}</h1>
        <button class="text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 ml-1" title="Edit class" @click="openEdit">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.232 5.232l3.536 3.536M9 11l6.536-6.536a2 2 0 112.828 2.828L11.828 13.828a2 2 0 01-.828.497l-3 .75.75-3a2 2 0 01.497-.828z" />
          </svg>
        </button>
      </div>
      <p class="text-sm text-slate-500 dark:text-slate-400 ml-7">
        <span v-if="cls.code">{{ cls.code }}</span>
        <span v-if="cls.code && cls.instructor"> · </span>
        <span v-if="cls.instructor">{{ cls.instructor }}</span>
        <span v-if="cls.schedule"> · {{ cls.schedule }}</span>
      </p>
    </div>

    <!-- Tab switcher + actions -->
    <div class="flex items-center justify-between mb-5">
      <div class="flex rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden text-sm">
        <button
          class="px-4 py-1.5 font-medium transition-colors"
          :class="view === 'list' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
          @click="setView('list')"
        >List</button>
        <button
          class="px-4 py-1.5 font-medium transition-colors border-l border-slate-200 dark:border-slate-700"
          :class="view === 'files' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
          @click="setView('files')"
        >Files</button>
        <button
          class="px-4 py-1.5 font-medium transition-colors border-l border-slate-200 dark:border-slate-700"
          :class="view === 'calendar' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
          @click="setView('calendar')"
        >Calendar</button>
        <button
          class="px-4 py-1.5 font-medium transition-colors border-l border-slate-200 dark:border-slate-700"
          :class="view === 'grades' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
          @click="setView('grades')"
        >Grades</button>
      </div>

      <div class="flex items-center gap-3">
        <button
          v-if="view === 'calendar'"
          class="text-sm text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 font-medium flex items-center gap-1.5"
          @click="showScheduleEditor = true"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          Edit schedule
        </button>

        <!-- View-format selector -->
        <div
          v-if="view === 'list' && sources.length > 0"
          class="flex rounded-md border border-slate-200 dark:border-slate-700 overflow-hidden"
        >
          <button
            type="button"
            title="Rows"
            class="px-2 py-1.5 transition-colors"
            :class="viewFormat === 'rows' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            @click="setFormat('rows')"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button
            type="button"
            title="Compact"
            class="px-2 py-1.5 border-l border-slate-200 dark:border-slate-700 transition-colors"
            :class="viewFormat === 'compact' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            @click="setFormat('compact')"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7h16M4 12h16M4 17h16" />
            </svg>
          </button>
          <button
            type="button"
            title="Tiles"
            class="px-2 py-1.5 border-l border-slate-200 dark:border-slate-700 transition-colors"
            :class="viewFormat === 'tiles' ? 'bg-slate-900 text-white dark:bg-slate-700' : 'bg-white dark:bg-slate-800 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            @click="setFormat('tiles')"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 5h6v6H4zM14 5h6v6h-6zM4 13h6v6H4zM14 13h6v6h-6z" />
            </svg>
          </button>
        </div>

        <button
          v-if="view === 'list'"
          class="flex items-center gap-1.5 text-sm text-indigo-600 hover:text-indigo-800 font-medium"
          @click="showNewSource = true"
        >
          <span class="text-base leading-none">+</span> Add Source
        </button>
      </div>
    </div>

    <!-- LIST VIEW -->
    <template v-if="view === 'list'">
      <!-- Filter bar (shown when class has 5+ sources) -->
      <div
        v-if="!loading && sources.length >= 5"
        class="mb-4 flex items-center gap-3 flex-wrap"
      >
        <div class="relative w-72 max-w-full">
          <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="7" /><path stroke-linecap="round" d="M20 20l-3.5-3.5" />
          </svg>
          <input
            ref="searchInputEl"
            v-model="searchQuery"
            type="text"
            placeholder="Find a source…"
            class="input pl-8"
            @keydown.esc.prevent="clearSearch"
          />
        </div>
        <div class="flex gap-1.5 flex-wrap">
          <button
            v-for="t in typeFilterOptions"
            :key="t"
            type="button"
            class="px-2.5 py-1 text-xs font-medium rounded-full transition-colors"
            :class="typeFilter === t
              ? 'bg-indigo-600 text-white'
              : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700 hover:border-indigo-300'"
            @click="typeFilter = t"
          >{{ t }}</button>
        </div>
      </div>

      <div v-if="loading" class="text-slate-400 text-sm">Loading…</div>

      <div v-else-if="sources.length === 0" class="text-center py-16 text-slate-400">
        <p class="text-4xl mb-3">📄</p>
        <p class="text-sm">No sources yet. Add a lecture, reading, or other source.</p>
      </div>

      <div v-else-if="filteredSources.length === 0" class="text-center py-12 text-slate-400">
        <p class="text-sm">No sources match your filter.</p>
        <button class="text-indigo-500 hover:text-indigo-700 text-sm font-medium mt-2" @click="clearFilter">Clear filter</button>
      </div>

      <!-- Rows -->
      <div v-else-if="viewFormat === 'rows'" ref="listEl" class="space-y-2">
        <div
          v-for="src in filteredSources"
          :key="src.id"
          class="group flex items-center gap-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-4 py-3 hover:border-indigo-300 dark:hover:border-indigo-500 transition-colors cursor-pointer"
          @click="$router.push(`/sources/${src.id}`)"
        >
          <div class="drag-handle text-slate-300 group-hover:text-slate-400 cursor-grab active:cursor-grabbing shrink-0" @click.stop>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <circle cx="9" cy="6" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="9" cy="18" r="1.5"/>
              <circle cx="15" cy="6" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="15" cy="18" r="1.5"/>
            </svg>
          </div>
          <!-- Completion circle (only for items with a due date) -->
          <button
            v-if="src.due_date"
            type="button"
            class="shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors"
            :class="src.completed
              ? 'border-green-500 bg-green-500 hover:bg-transparent hover:border-slate-300'
              : 'border-slate-300 dark:border-slate-600 hover:border-green-500'"
            :title="src.completed ? 'Mark as incomplete' : 'Mark as complete'"
            @click.stop="src.completed ? setCompleted(src.id, false) : requestComplete(src.id)"
          >
            <svg v-if="src.completed" class="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </button>
          <span class="shrink-0 text-xs font-medium px-2 py-0.5 rounded-full" :class="typeBadgeClass(src.type)">
            {{ src.type }}
          </span>
          <PriorityBadge :priority="src.priority" size="xs" @click.stop @change="setPriority(src.id, $event)" />
          <div class="flex-1 min-w-0">
            <p class="font-medium text-slate-800 dark:text-slate-200 truncate">{{ src.title }}</p>
            <p v-if="src.reference" class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ src.reference }}</p>
          </div>
          <div class="shrink-0 flex items-center gap-3">
            <span v-if="src.due_date" class="text-xs font-medium" :class="src.completed ? 'text-green-500' : 'text-rose-500'">
              {{ src.completed ? '✓ Done' : 'Due ' + formatDate(src.due_date) }}
            </span>
            <span v-if="src.date" class="text-xs text-slate-400">{{ formatDate(src.date) }}</span>
          </div>
          <button
            class="shrink-0 text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
            title="Delete source"
            @click.stop="deleteSource(src.id)"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Compact -->
      <div v-else-if="viewFormat === 'compact'" ref="listEl" class="space-y-1">
        <div
          v-for="src in filteredSources"
          :key="src.id"
          class="group flex items-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-3 py-1.5 hover:border-indigo-300 dark:hover:border-indigo-500 transition-colors cursor-grab active:cursor-grabbing"
          @click="$router.push(`/sources/${src.id}`)"
        >
          <button
            v-if="src.due_date"
            type="button"
            class="shrink-0 w-3.5 h-3.5 rounded-full border-2 flex items-center justify-center transition-colors"
            :class="src.completed
              ? 'border-green-500 bg-green-500 hover:bg-transparent hover:border-slate-300'
              : 'border-slate-300 dark:border-slate-600 hover:border-green-500'"
            :title="src.completed ? 'Mark as incomplete' : 'Mark as complete'"
            @click.stop="src.completed ? setCompleted(src.id, false) : requestComplete(src.id)"
          >
            <svg v-if="src.completed" class="w-2 h-2 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </button>
          <span class="w-2 h-2 rounded-full shrink-0" :class="typeDotClass(src.type)" :title="src.type" />
          <PriorityBadge :priority="src.priority" size="xs" @click.stop @change="setPriority(src.id, $event)" />
          <span class="text-sm truncate flex-1 text-slate-800 dark:text-slate-200">{{ src.title }}</span>
          <span v-if="src.due_date" class="shrink-0 text-xs font-medium" :class="src.completed ? 'text-green-500' : 'text-rose-500'">
            {{ src.completed ? '✓ Done' : 'Due ' + formatDate(src.due_date) }}
          </span>
          <span v-if="src.date" class="shrink-0 text-xs text-slate-400">{{ formatDate(src.date) }}</span>
          <button
            class="shrink-0 text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
            title="Delete source"
            @click.stop="deleteSource(src.id)"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Tiles -->
      <div v-else ref="listEl" class="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
        <div
          v-for="src in filteredSources"
          :key="src.id"
          class="group bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden hover:border-indigo-300 dark:hover:border-indigo-500 transition-colors cursor-pointer flex flex-col"
          @click="$router.push(`/sources/${src.id}`)"
        >
          <div class="relative aspect-[3/4] bg-slate-50 dark:bg-slate-900">
            <DeepReadCover
              v-if="sourcePdfCover.get(src.id)"
              :url="sourcePdfCover.get(src.id) as string"
            />
            <div
              v-else
              class="w-full h-full flex items-center justify-center"
              :class="typeBlockClass(src.type)"
            >
              <span class="text-white text-base font-semibold opacity-70 px-3 text-center">{{ src.type }}</span>
            </div>
            <button
              class="absolute top-1.5 right-1.5 p-1 rounded bg-white/85 dark:bg-slate-800/85 text-slate-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
              title="Delete source"
              @click.stop="deleteSource(src.id)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="p-2 flex-1 min-w-0">
            <div class="flex items-center gap-1 mb-0.5">
              <PriorityBadge v-if="src.priority !== 0" :priority="src.priority" size="xs" :editable="false" />
            </div>
            <p class="font-medium text-xs text-slate-800 dark:text-slate-200 truncate">{{ src.title }}</p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span v-if="src.due_date" class="text-[10px] text-rose-500 font-medium truncate">Due {{ formatDate(src.due_date) }}</span>
              <span v-if="src.date" class="text-[10px] text-slate-400 truncate">{{ formatDate(src.date) }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- CALENDAR VIEW -->
    <template v-else-if="view === 'calendar'">
      <CalendarView
        :events="calendarEvents"
        :loading="calendarLoading"
        :initial-date="currentMonth"
        @month-change="onMonthChange"
        @event-click="popoverEvent = $event"
      />
    </template>

    <!-- FILES VIEW -->
    <template v-else-if="view === 'files'">
      <AttachmentPanel
        :scope="{ type: 'class', id: classID }"
        :focus-context="classFocusContext"
        :return-to="$route.path"
      />
    </template>

    <!-- GRADES VIEW -->
    <template v-else-if="view === 'grades'">
      <GradePanel :class-i-d="classID" />
    </template>

    <!-- New Source Modal -->
    <BaseModal :open="showNewSource" title="Add Source" @close="showNewSource = false">
      <form class="space-y-4" @submit.prevent="createSource">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Type</label>
          <select v-model="srcForm.type" class="input">
            <option v-for="t in sourceTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Title *</label>
          <input v-model="srcForm.title" required class="input" placeholder="e.g. Lecture 3: Divide & Conquer" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Date</label>
            <input v-model="srcForm.date" type="date" class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Due date</label>
            <input v-model="srcForm.due_date" type="date" class="input" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Reference</label>
          <input v-model="srcForm.reference" class="input" placeholder="CLRS Ch. 4" />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Priority</label>
          <PriorityBadge :priority="srcForm.priority" @change="srcForm.priority = $event" />
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showNewSource = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="!srcForm.title">Add</button>
        </div>
      </form>
    </BaseModal>

    <!-- Edit Class Modal -->
    <BaseModal :open="showEdit" title="Edit Class" @close="showEdit = false">
      <form class="space-y-4" @submit.prevent="saveEdit">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Class name *</label>
          <input v-model="editForm.name" required class="input" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Course code</label>
            <input v-model="editForm.code" class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Color</label>
            <input v-model="editForm.color" type="color" class="h-9 w-full rounded-md border border-slate-300 cursor-pointer" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Instructor</label>
          <input v-model="editForm.instructor" class="input" />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Schedule (display label)</label>
          <input v-model="editForm.schedule" class="input" placeholder="MWF 10–11am" />
        </div>
        <div class="flex justify-between pt-2">
          <button type="button" class="text-red-500 hover:text-red-700 text-sm font-medium" @click="deleteClass">Delete class</button>
          <div class="flex gap-2">
            <button type="button" class="btn-ghost" @click="showEdit = false">Cancel</button>
            <button type="submit" class="btn-primary">Save</button>
          </div>
        </div>
      </form>
    </BaseModal>

    <!-- Schedule Editor Modal -->
    <BaseModal :open="showScheduleEditor" title="Class Schedule" @close="showScheduleEditor = false">
      <ClassMeetingsEditor v-if="showScheduleEditor" :class-i-d="classID" @change="refreshCalendar" />
    </BaseModal>

    <!-- Event Popover -->
    <CalendarEventPopover :event="popoverEvent" @close="popoverEvent = null" />

    <!-- Delete confirmation -->
    <ConfirmModal
      :open="confirmDelete.open"
      :title="confirmDelete.title"
      :message="confirmDelete.message"
      @confirm="runConfirm"
      @cancel="confirmDelete.open = false"
    />

    <!-- Complete confirmation -->
    <ConfirmModal
      :open="pendingComplete !== null"
      title="Mark as complete?"
      message="Mark this source as completed. You can undo this by clicking the green circle again."
      confirm-label="Mark complete"
      :danger="false"
      @confirm="confirmComplete"
      @cancel="pendingComplete = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Sortable from 'sortablejs'
import { api, type Source, type Class, type CalendarEventProps } from '@/api'
import { useClassStore } from '@/stores/class'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import PriorityBadge from '@/components/PriorityBadge.vue'
import CalendarView from '@/components/CalendarView.vue'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import CalendarEventPopover from '@/components/CalendarEventPopover.vue'
import ClassMeetingsEditor from '@/components/ClassMeetingsEditor.vue'
import DeepReadCover from '@/components/DeepReadCover.vue'
import GradePanel from '@/components/GradePanel.vue'
import type { CalendarEvent } from '@/api'

const route = useRoute()
const router = useRouter()
const classStore = useClassStore()

const classID = Number(route.params.id)
const cls = ref<Class | null>(null)
const sources = ref<Source[]>([])
const loading = ref(true)
const listEl = ref<HTMLElement | null>(null)
let sortable: Sortable | null = null

// View state — synced to ?view= query param
type ClassViewTab = 'list' | 'calendar' | 'files' | 'grades'
const view = ref<ClassViewTab>(
  (['calendar', 'files', 'grades'] as const).includes(route.query.view as ClassViewTab)
    ? (route.query.view as ClassViewTab)
    : 'list'
)

function setView(v: ClassViewTab) {
  view.value = v
  router.replace({ query: { ...route.query, view: v === 'list' ? undefined : v } })
}

// View-format (rows / compact / tiles) — persisted per-class in localStorage
type SourceViewFormat = 'rows' | 'compact' | 'tiles'
const VIEW_FORMAT_KEY = `class:${classID}:source-view`
const viewFormat = ref<SourceViewFormat>(loadFormat())

function loadFormat(): SourceViewFormat {
  try {
    const v = localStorage.getItem(VIEW_FORMAT_KEY)
    if (v === 'rows' || v === 'compact' || v === 'tiles') return v
  } catch { /* ignore storage failures */ }
  return 'rows'
}

function setFormat(f: SourceViewFormat) {
  if (viewFormat.value === f) return
  viewFormat.value = f
  try { localStorage.setItem(VIEW_FORMAT_KEY, f) } catch { /* ignore */ }
}

// Type → color helpers (Tailwind class strings) for compact dots and tile blocks
const TYPE_COLOR: Record<string, string> = {
  Lecture: 'bg-blue-500',
  Reading: 'bg-purple-500',
  'Problem Set': 'bg-orange-500',
  Discussion: 'bg-green-500',
  Lab: 'bg-teal-500',
  Other: 'bg-slate-400'
}
function typeDotClass(type: string) { return TYPE_COLOR[type] ?? 'bg-slate-400' }
function typeBlockClass(type: string) { return TYPE_COLOR[type] ?? 'bg-slate-400' }

// In-class search + type filter (filtering is purely client-side; sources already loaded)
const searchQuery = ref('')
const typeFilter = ref<string>('All')
const searchInputEl = ref<HTMLInputElement | null>(null)
const typeFilterOptions = ['All', 'Lecture', 'Reading', 'Problem Set', 'Discussion', 'Lab', 'Other']

const filteredSources = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  const t = typeFilter.value
  if (!q && t === 'All') return sources.value
  return sources.value.filter((s) => {
    const matchesQ = !q || s.title.toLowerCase().includes(q)
    const matchesT = t === 'All' || s.type === t
    return matchesQ && matchesT
  })
})
const isFilterActive = computed(() => searchQuery.value.trim() !== '' || typeFilter.value !== 'All')

function clearSearch() {
  searchQuery.value = ''
  searchInputEl.value?.blur()
}
function clearFilter() {
  searchQuery.value = ''
  typeFilter.value = 'All'
}

// Search clears on view-format change (per spec); type filter persists.
watch(viewFormat, () => { searchQuery.value = '' })

// Lazy PDF-cover lookup for tile mode: source.id → file URL or null (no PDF)
const sourcePdfCover = ref(new Map<number, string | null>())
async function ensurePdfCovers() {
  const todo = sources.value.filter(s => !sourcePdfCover.value.has(s.id))
  await Promise.all(todo.map(async (s) => {
    try {
      const atts = await api.attachments.listForSource(s.id)
      const pdf = atts.find(a => a.mime_type === 'application/pdf')
      sourcePdfCover.value.set(s.id, pdf ? api.attachments.fileUrl(pdf.id, 'inline') : null)
    } catch {
      sourcePdfCover.value.set(s.id, null)
    }
  }))
}

const classFocusContext = computed(() => {
  if (!cls.value) return undefined
  return {
    class_id: cls.value.id,
    class_name: cls.value.name,
    class_color: cls.value.color,
    source_id: null,
    source_title: ''
  }
})

// Confirm-delete modal state
const confirmDelete = ref<{ open: boolean; title: string; message: string; action: (() => Promise<void>) | null }>({
  open: false, title: '', message: '', action: null
})
function askConfirm(title: string, message: string, action: () => Promise<void>) {
  confirmDelete.value = { open: true, title, message, action }
}
async function runConfirm() {
  const action = confirmDelete.value.action
  confirmDelete.value = { open: false, title: '', message: '', action: null }
  if (action) await action()
}

// Calendar state
const calendarEvents = ref<CalendarEvent[]>([])
const calendarLoading = ref(false)
const currentMonth = ref(todayMonth())
const popoverEvent = ref<(CalendarEventProps & { title: string; start: string }) | null>(null)

function todayMonth() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

async function fetchCalendarEvents(month: string) {
  calendarLoading.value = true
  try {
    calendarEvents.value = await api.calendar.forClass(classID, month)
  } finally {
    calendarLoading.value = false
  }
}

async function onMonthChange(month: string) {
  if (month !== currentMonth.value) {
    currentMonth.value = month
    await fetchCalendarEvents(month)
  }
}

async function refreshCalendar() {
  await fetchCalendarEvents(currentMonth.value)
}

watch(view, async (v) => {
  if (v === 'calendar' && calendarEvents.value.length === 0) {
    await fetchCalendarEvents(currentMonth.value)
  }
})

// Modals
const showNewSource = ref(false)
const showEdit = ref(false)
const showScheduleEditor = ref(false)
const sourceTypes = ['Lecture', 'Reading', 'Problem Set', 'Discussion', 'Lab', 'Other']
function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

const srcForm = ref({ type: 'Lecture', title: '', date: todayISO(), due_date: '', reference: '', priority: 0 })
const editForm = ref({ name: '', code: '', instructor: '', schedule: '', color: '#6366f1' })

// Sortable lifecycle — re-attaches when view/format changes; disabled in tiles
function destroySortable() {
  if (sortable) { sortable.destroy(); sortable = null }
}
function attachSortable(el: HTMLElement) {
  destroySortable()
  sortable = new Sortable(el, {
    handle: viewFormat.value === 'rows' ? '.drag-handle' : undefined,
    animation: 150,
    onEnd: async (evt) => {
      const oldIdx = evt.oldIndex ?? 0
      const newIdx = evt.newIndex ?? 0
      const moved = sources.value.splice(oldIdx, 1)[0]
      sources.value.splice(newIdx, 0, moved)
      await api.sources.reorder(classID, sources.value.map(s => s.id))
    }
  })
}
async function refreshSortable() {
  await nextTick()
  // Drag-to-reorder is disabled in tiles, when a filter is active (indices wouldn't match), or when list is empty.
  if (view.value !== 'list' || viewFormat.value === 'tiles' || isFilterActive.value || sources.value.length === 0 || !listEl.value) {
    destroySortable()
    return
  }
  attachSortable(listEl.value)
}

watch([view, viewFormat, () => sources.value.length, isFilterActive], () => {
  refreshSortable()
  if (view.value === 'list' && viewFormat.value === 'tiles') ensurePdfCovers()
})

// Keyboard: '/' focuses search; 1/2/3 cycles list-view formats (when ClassView's list view is active).
function onKey(e: KeyboardEvent) {
  if (view.value !== 'list') return
  const t = e.target as HTMLElement | null
  const inInput = !!t && (t.tagName.toLowerCase() === 'input' || t.tagName.toLowerCase() === 'textarea' || t.isContentEditable)

  if (e.key === '/' && !inInput && !e.metaKey && !e.ctrlKey && !e.altKey && !e.shiftKey) {
    if (searchInputEl.value) {
      e.preventDefault()
      searchInputEl.value.focus()
    }
    return
  }

  if (inInput) return
  if (e.metaKey || e.ctrlKey || e.altKey || e.shiftKey) return
  if (e.key === '1') { setFormat('rows'); e.preventDefault() }
  else if (e.key === '2') { setFormat('compact'); e.preventDefault() }
  else if (e.key === '3') { setFormat('tiles'); e.preventDefault() }
}

onMounted(async () => {
  ;[cls.value, sources.value] = await Promise.all([
    api.classes.get(classID),
    api.sources.list(classID)
  ])
  loading.value = false

  // Pre-fetch calendar if starting in calendar view
  if (view.value === 'calendar') {
    await fetchCalendarEvents(currentMonth.value)
  }

  await refreshSortable()
  if (view.value === 'list' && viewFormat.value === 'tiles') await ensurePdfCovers()

  window.addEventListener('keydown', onKey)
})

onUnmounted(() => {
  destroySortable()
  window.removeEventListener('keydown', onKey)
})

function openEdit() {
  if (!cls.value) return
  editForm.value = { name: cls.value.name, code: cls.value.code, instructor: cls.value.instructor, schedule: cls.value.schedule, color: cls.value.color }
  showEdit.value = true
}

async function saveEdit() {
  const updated = await classStore.update(classID, editForm.value)
  cls.value = updated
  showEdit.value = false
}

function deleteClass() {
  askConfirm(
    'Delete class',
    'This will permanently delete this class, all its sources, notes, and attachments. This cannot be undone.',
    async () => { await classStore.remove(classID); router.push('/') }
  )
}

async function createSource() {
  const s = await api.sources.create(classID, srcForm.value)
  sources.value.push(s)
  srcForm.value = { type: 'Lecture', title: '', date: todayISO(), due_date: '', reference: '', priority: 0 }
  showNewSource.value = false
  if (view.value === 'calendar') await refreshCalendar()
  if (viewFormat.value === 'tiles') sourcePdfCover.value.set(s.id, null) // new source has no PDF yet
}

async function setPriority(id: number, priority: number) {
  const s = await api.sources.update(id, { priority })
  const idx = sources.value.findIndex(x => x.id === id)
  if (idx !== -1) sources.value[idx] = s
}

async function setCompleted(id: number, completed: boolean) {
  const s = await api.sources.update(id, { completed })
  const idx = sources.value.findIndex(x => x.id === id)
  if (idx !== -1) sources.value[idx] = s
}

const pendingComplete = ref<number | null>(null)
function requestComplete(id: number) { pendingComplete.value = id }
async function confirmComplete() {
  if (pendingComplete.value === null) return
  await setCompleted(pendingComplete.value, true)
  pendingComplete.value = null
}

function deleteSource(id: number) {
  askConfirm(
    'Delete source',
    'This will permanently delete this source and all its notes and attachments. This cannot be undone.',
    async () => {
      await api.sources.delete(id)
      sources.value = sources.value.filter(s => s.id !== id)
      sourcePdfCover.value.delete(id)
      if (view.value === 'calendar') await refreshCalendar()
    }
  )
}

function typeBadgeClass(type: string) {
  const map: Record<string, string> = {
    Lecture: 'bg-blue-100 text-blue-700', Reading: 'bg-purple-100 text-purple-700',
    'Problem Set': 'bg-orange-100 text-orange-700', Discussion: 'bg-green-100 text-green-700',
    Lab: 'bg-teal-100 text-teal-700', Other: 'bg-slate-100 text-slate-600'
  }
  return map[type] ?? 'bg-slate-100 text-slate-600'
}

function formatDate(d: string) {
  if (!d) return ''
  return new Date(d + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

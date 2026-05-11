import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error' | 'locked'

export const useEditorStore = defineStore('editor', () => {
  const body = ref('')
  const dirty = ref(false)
  const saveStatus = ref<SaveStatus>('idle')
  const locked = ref(false)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null
  let currentSourceID: number | null = null

  function setSourceID(id: number) {
    currentSourceID = id
    dirty.value = false
    locked.value = false
    saveStatus.value = 'idle'
  }

  function setLockedFromServer(v: boolean) {
    locked.value = v
    if (v) saveStatus.value = 'locked'
    else if (saveStatus.value === 'locked') saveStatus.value = 'idle'
  }

  function onChange(newBody: string) {
    if (locked.value) return // belt: client-side gate so we don't queue saves
    body.value = newBody
    dirty.value = true
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => save(), 800)
  }

  async function save(retries = 3): Promise<void> {
    if (!currentSourceID || !dirty.value || locked.value) return
    saveStatus.value = 'saving'
    try {
      await api.notes.upsert(currentSourceID, body.value)
      dirty.value = false
      saveStatus.value = 'saved'
      setTimeout(() => {
        if (saveStatus.value === 'saved') saveStatus.value = 'idle'
      }, 2000)
    } catch (e) {
      // Server-side lock — sync state, don't retry
      if (e instanceof Error && /locked/i.test(e.message)) {
        locked.value = true
        saveStatus.value = 'locked'
        dirty.value = false
        return
      }
      if (retries > 0) {
        setTimeout(() => save(retries - 1), 3000)
        saveStatus.value = 'error'
      }
    }
  }

  function flushSave() {
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    return save()
  }

  async function setLock(v: boolean) {
    if (!currentSourceID) return
    // Flush any pending save BEFORE we lock so the user's last edit lands.
    if (v) await flushSave()
    const note = await api.notes.setLock(currentSourceID, v)
    locked.value = !!note.locked
    saveStatus.value = locked.value ? 'locked' : 'idle'
  }

  return {
    body, dirty, saveStatus, locked,
    setSourceID, setLockedFromServer, onChange, save, flushSave, setLock
  }
})

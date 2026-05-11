import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Class } from '@/api'

export const useClassStore = defineStore('class', () => {
  const classes = ref<Class[]>([])
  const loading = ref(false)

  async function loadForSemester(semesterID: number) {
    loading.value = true
    try {
      classes.value = await api.classes.list(semesterID)
    } finally {
      loading.value = false
    }
  }

  async function create(semesterID: number, data: Partial<Class>) {
    const c = await api.classes.create(semesterID, data)
    classes.value.push(c)
    return c
  }

  async function update(id: number, data: Partial<Class>) {
    const c = await api.classes.update(id, data)
    const idx = classes.value.findIndex(x => x.id === id)
    if (idx !== -1) classes.value[idx] = c
    return c
  }

  async function remove(id: number) {
    await api.classes.delete(id)
    classes.value = classes.value.filter(c => c.id !== id)
  }

  return { classes, loading, loadForSemester, create, update, remove }
})

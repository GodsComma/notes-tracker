import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type Semester } from '@/api'

export const useSemesterStore = defineStore('semester', () => {
  const semesters = ref<Semester[]>([])
  const activeSemesterID = ref<number | null>(null)

  const activeSemester = computed(() =>
    semesters.value.find(s => s.id === activeSemesterID.value) ?? null
  )

  async function load() {
    semesters.value = await api.semesters.list()
    // Default to the first (most recent) semester
    if (!activeSemesterID.value && semesters.value.length > 0) {
      activeSemesterID.value = semesters.value[0].id
    }
  }

  async function create(name: string, startDate: string, endDate: string) {
    const s = await api.semesters.create({ name, start_date: startDate, end_date: endDate })
    semesters.value.unshift(s)
    activeSemesterID.value = s.id
    return s
  }

  async function update(id: number, name: string, startDate: string, endDate: string) {
    const s = await api.semesters.update(id, { name, start_date: startDate, end_date: endDate })
    const idx = semesters.value.findIndex(x => x.id === id)
    if (idx !== -1) semesters.value[idx] = s
    return s
  }

  async function remove(id: number) {
    await api.semesters.delete(id)
    semesters.value = semesters.value.filter(s => s.id !== id)
    if (activeSemesterID.value === id) {
      activeSemesterID.value = semesters.value[0]?.id ?? null
    }
  }

  function setActive(id: number) {
    activeSemesterID.value = id
  }

  return { semesters, activeSemesterID, activeSemester, load, create, update, remove, setActive }
})

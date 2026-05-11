// Typed API wrapper — all fetch calls live here

export interface Semester {
  id: number
  name: string
  start_date: string
  end_date: string
  created_at: string
}

export interface Class {
  id: number
  semester_id: number
  name: string
  code: string
  instructor: string
  schedule: string
  color: string
}

export interface Source {
  id: number
  class_id: number
  type: string
  title: string
  date: string
  due_date: string
  reference: string
  sort_order: number
  priority: number   // 0=None, 1=High, 2=Medium, 3=Low
  completed: boolean
}

export interface SourceWithClass extends Source {
  class_name: string
  class_color: string
}

export interface ClassMeeting {
  id: number
  class_id: number
  day_of_week: number
  start_time: string
  end_time: string
  location: string
}

export interface CalendarEventProps {
  type: 'source' | 'deadline' | 'meeting'
  sourceID?: number
  classID?: number
  className?: string
  sourceType?: string
}

export interface CalendarEvent {
  id: string
  title: string
  start: string
  end?: string
  allDay: boolean
  color: string
  extendedProps: CalendarEventProps
}

export interface Note {
  id?: number
  source_id: number
  body: string
  locked?: boolean
  created_at?: string
  updated_at?: string
}

export interface Attachment {
  id: number
  class_id: number | null
  source_id: number | null
  filename: string
  mime_type: string
  size_bytes: number
  created_at: string
}

export interface PDFLibraryItem {
  id: number
  filename: string
  size_bytes: number
  created_at: string
  class_id: number
  class_name: string
  class_color: string
  source_id: number | null
  source_title: string
}

export interface FocusSession {
  id: number
  date: string
  duration_sec: number
  attachment_id: number
  filename: string
  completed: boolean
  created_at: string
}

export interface SearchResult {
  note_id: number
  source_id: number
  source_title: string
  source_date: string
  class_id: number
  class_name: string
  class_color: string
  snippet: string
}

export interface GradeCategory {
  id: number
  rubric_id: number
  name: string
  weight: number
  is_extra_credit: boolean
  earned: number | null
  possible: number
  sort_order: number
}

export interface GradeRubric {
  id: number
  class_id: number
  grading_type: 'letter' | 'pass_fail'
  plus_minus: boolean
  thresholds: string
  categories: GradeCategory[]
  created_at: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const isFormData = init?.body instanceof FormData
  const res = await fetch(`/api/v1/${path}`, {
    ...init,
    credentials: 'include',
    headers: isFormData
      ? (init?.headers ?? {})
      : { 'Content-Type': 'application/json', ...(init?.headers ?? {}) }
  })
  if (res.status === 401) {
    // Token expired or missing — redirect to login without import cycle
    window.location.href = '/login'
    return undefined as T
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// --- Semesters ---
export const api = {
  semesters: {
    list: () => request<Semester[]>('/semesters'),
    create: (data: Pick<Semester, 'name' | 'start_date' | 'end_date'>) =>
      request<Semester>('/semesters', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Semester>) =>
      request<Semester>(`/semesters/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<void>(`/semesters/${id}`, { method: 'DELETE' })
  },

  classes: {
    get: (id: number) => request<Class>(`/classes/${id}`),
    list: (semesterID: number) => request<Class[]>(`/semesters/${semesterID}/classes`),
    create: (semesterID: number, data: Partial<Class>) =>
      request<Class>(`/semesters/${semesterID}/classes`, { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Class>) =>
      request<Class>(`/classes/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<void>(`/classes/${id}`, { method: 'DELETE' })
  },

  sources: {
    get: (id: number) => request<Source>(`/sources/${id}`),
    list: (classID: number) => request<Source[]>(`/classes/${classID}/sources`),
    create: (classID: number, data: Partial<Source>) =>
      request<Source>(`/classes/${classID}/sources`, { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Source>) =>
      request<Source>(`/sources/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    listForSemester: (semesterID: number) =>
      request<SourceWithClass[]>(`/semesters/${semesterID}/sources`),
    delete: (id: number) =>
      request<void>(`/sources/${id}`, { method: 'DELETE' }),
    reorder: (classID: number, ids: number[]) =>
      request<void>(`/classes/${classID}/sources/reorder`, { method: 'PATCH', body: JSON.stringify({ ids }) })
  },

  notes: {
    get: (sourceID: number) => request<Note>(`/sources/${sourceID}/note`),
    upsert: (sourceID: number, body: string) =>
      request<Note>(`/sources/${sourceID}/note`, { method: 'PUT', body: JSON.stringify({ body }) }),
    setLock: (sourceID: number, locked: boolean) =>
      request<Note>(`/sources/${sourceID}/note/lock`, { method: 'PATCH', body: JSON.stringify({ locked }) })
  },

  calendar: {
    forClass: (classID: number, month: string) =>
      request<CalendarEvent[]>(`/classes/${classID}/calendar?month=${month}`),
    forSemester: (semesterID: number, month: string) =>
      request<CalendarEvent[]>(`/semesters/${semesterID}/calendar?month=${month}`)
  },

  meetings: {
    list: (classID: number) => request<ClassMeeting[]>(`/classes/${classID}/meetings`),
    create: (classID: number, data: Omit<ClassMeeting, 'id' | 'class_id'>) =>
      request<ClassMeeting>(`/classes/${classID}/meetings`, { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Partial<ClassMeeting>) =>
      request<ClassMeeting>(`/meetings/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<void>(`/meetings/${id}`, { method: 'DELETE' })
  },

  attachments: {
    listForClass: (classID: number) =>
      request<Attachment[]>(`/classes/${classID}/attachments`),
    listForSource: (sourceID: number) =>
      request<Attachment[]>(`/sources/${sourceID}/attachments`),
    uploadToClass: (classID: number, file: File) => {
      const form = new FormData()
      form.append('file', file)
      return request<Attachment>(`/classes/${classID}/attachments`, { method: 'POST', body: form })
    },
    uploadToSource: (sourceID: number, file: File) => {
      const form = new FormData()
      form.append('file', file)
      return request<Attachment>(`/sources/${sourceID}/attachments`, { method: 'POST', body: form })
    },
    fileUrl: (id: number, disposition: 'inline' | 'attachment' = 'inline') =>
      `/api/v1/attachments/${id}/file?disposition=${disposition}`,
    getProgress: (id: number) =>
      request<{ last_page: number }>(`/attachments/${id}/progress`),
    saveProgress: (id: number, lastPage: number) =>
      request<void>(`/attachments/${id}/progress`, {
        method: 'PUT',
        body: JSON.stringify({ last_page: lastPage })
      }),
    delete: (id: number) =>
      request<void>(`/attachments/${id}`, { method: 'DELETE' })
  },

  library: {
    pdfs: () => request<PDFLibraryItem[]>('/library/pdfs')
  },

  focusSessions: {
    list: () => request<FocusSession[]>('/focus-sessions'),
    log: (data: Omit<FocusSession, 'id' | 'created_at'>) =>
      request<FocusSession>('/focus-sessions', { method: 'POST', body: JSON.stringify(data) })
  },

  search: (q: string, classID?: number) => {
    const params = new URLSearchParams({ q })
    if (classID) params.set('class_id', String(classID))
    return request<SearchResult[]>(`/search?${params}`)
  },

  grades: {
    getRubric: (classID: number) => request<GradeRubric>(`/classes/${classID}/rubric`),
    putRubric: (classID: number, data: Pick<GradeRubric, 'grading_type' | 'plus_minus' | 'thresholds'>) =>
      request<GradeRubric>(`/classes/${classID}/rubric`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteRubric: (classID: number) =>
      request<void>(`/classes/${classID}/rubric`, { method: 'DELETE' }),
    addCategory: (rubricID: number, data: Omit<GradeCategory, 'id' | 'rubric_id' | 'sort_order' | 'earned'>) =>
      request<GradeCategory>(`/rubrics/${rubricID}/categories`, { method: 'POST', body: JSON.stringify(data) }),
    updateCategory: (id: number, data: Partial<GradeCategory>) =>
      request<GradeCategory>(`/grade-categories/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    deleteCategory: (id: number) =>
      request<void>(`/grade-categories/${id}`, { method: 'DELETE' })
  }
}

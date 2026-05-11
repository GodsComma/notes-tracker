export function fileIcon(mimeType: string): string {
  if (mimeType.startsWith('image/')) return '🖼'
  if (mimeType === 'application/pdf') return '📄'
  if (mimeType.startsWith('video/')) return '🎬'
  if (mimeType.startsWith('audio/')) return '🎵'
  if (mimeType.includes('zip') || mimeType.includes('compressed')) return '🗜'
  return '📎'
}

export function isPreviewable(mimeType: string): boolean {
  return mimeType.startsWith('image/') || mimeType === 'application/pdf'
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 ** 2).toFixed(1)} MB`
}

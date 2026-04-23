export function formatDate(input?: string | null): string {
  if (!input) return '—'
  const d = new Date(input)
  if (Number.isNaN(d.getTime())) return '—'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function formatDateTime(input?: string | null): string {
  if (!input) return '—'
  const d = new Date(input)
  if (Number.isNaN(d.getTime())) return '—'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}

export function formatRelative(input?: string | null, now = Date.now()): string {
  if (!input) return '—'
  const t = new Date(input).getTime()
  if (Number.isNaN(t)) return '—'
  const diff = Math.round((t - now) / 1000)
  const abs = Math.abs(diff)
  const units: [number, string][] = [
    [60, '秒'],
    [60, '分钟'],
    [24, '小时'],
    [30, '天'],
    [12, '月'],
    [Number.POSITIVE_INFINITY, '年'],
  ]
  let value = abs
  let unit = ''
  for (const [size, name] of units) {
    if (value < size) {
      unit = name
      break
    }
    value = Math.floor(value / size)
  }
  if (!unit) unit = '年'
  return diff < 0 ? `${value} ${unit}前` : `${value} ${unit}后`
}

export function formatReward(amount: number, currency: string): string {
  const rounded = Number.isInteger(amount) ? amount : amount.toFixed(2)
  return `${rounded} ${currency}`
}

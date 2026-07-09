export const VALIDATION = {
  username: { min: 4, max: 25 },
  password: { min: 7, max: 40 },
  first_name: { max: 30 },
  last_name: { max: 30 },
  bio: { max: 1000 },
  post_title: { min: 5, max: 300 },
  post_body: { min: 1, max: 7000 },
  comment_body: { max: 3000 },
  categories: { max: 5 },
} as const

export function validate_required(value: string, label: string): string | null {
  if (!value.trim()) return `${label} is required`
  return null
}

export function validate_length(
  value: string,
  label: string,
  min?: number,
  max?: number,
): string | null {
  if (min && value.length < min) return `${label} must be at least ${min} characters`
  if (max && value.length > max) return `${label} must be at most ${max} characters`
  return null
}

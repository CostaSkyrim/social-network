export const VALIDATION = {
  password: { min: 7, max: 40 },
} as const

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

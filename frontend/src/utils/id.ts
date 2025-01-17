/**
 * Helper function to get the ID from an object that might have either uppercase or lowercase 'id'
 * @param obj Object that might have ID or id property
 * @returns The ID value or undefined if neither exists
 */
export function getId(obj: any): string | undefined {
  if (!obj) return undefined
  return obj.ID || obj.id
} 
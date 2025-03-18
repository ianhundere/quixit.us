export function formatDate(date: string | Date): string {
    return new Date(date).toLocaleString();
}

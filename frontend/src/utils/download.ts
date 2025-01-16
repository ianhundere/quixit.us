export async function downloadFile(url: string): Promise<Blob> {
    const token = localStorage.getItem('access_token')
    const response = await fetch(url, {
        headers: {
            Authorization: `Bearer ${token}`
        }
    })
    if (!response.ok) throw new Error('Download failed')
    return response.blob()
} 
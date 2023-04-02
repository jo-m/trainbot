const blobsBaseURL = import.meta.env.VITE_BLOBS_URL

export function getBlobURL(blobName: string): string {
  return blobsBaseURL.trimRight('/') + '/' + blobName
}

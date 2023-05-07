const blobsBaseURL = import.meta.env.VITE_BLOBS_URL

export function getBlobURL(blobName: string): string {
  return blobsBaseURL.trimRight('/') + '/' + blobName
}

export function getBlobThumbURL(blobName: string): string {
  const thumbName = blobName.replace(/.jpg$/, '.thumb.jpg')
  return getBlobURL(thumbName)
}

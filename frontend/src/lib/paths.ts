import type { DateTime } from 'luxon'

const blobsBaseURL = import.meta.env.VITE_BLOBS_URL

// This matches time.Format() with '20060102_150405.999_Z07:00' from Go.
function formatFileTs(ts: DateTime): string {
  const base = ts.toFormat('yyyyMMdd_HHmmss')
  const frac = ts.toFormat('.SSS').replace(/\.?0*$/, '')
  const zone = ts.toFormat('ZZ').replace('+00:00', 'Z')

  return `${base}${frac}_${zone}`
}

export function imgFileName(ts: DateTime): string {
  return `train_${formatFileTs(ts)}.jpg`
}

export function gifFileName(ts: DateTime): string {
  return `train_${formatFileTs(ts)}.gif`
}

export function getBlobURL(blobName: string): string {
  return blobsBaseURL.trimRight('/') + '/' + blobName
}

export function getBlobThumbURL(blobName: string): string {
  const thumbName = blobName.replace(/.jpg$/, '.thumb.jpg')
  return getBlobURL(thumbName)
}

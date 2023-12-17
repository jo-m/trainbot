#include "c.h"

static const int four = 4;

static int64_t iabs(int64_t a) {
  if (a < 0) {
    return -a;
  }
  return a;
}

void RGBAC(const int m, const int n, const int s,
           /* pixels */
           const uint8_t* const pix,
           /* return parameters */
           retData* ret) {
  uint64_t sum[3] = {0};

  for (int y = 0; y < n; y++) {
    const int ys = y * s;
    for (int x = 0; x < m; x++) {
      const int ix = ys + x * four;
      sum[0] += (uint64_t)(pix[ix + 0]);
      sum[1] += (uint64_t)(pix[ix + 1]);
      sum[2] += (uint64_t)(pix[ix + 2]);
    }
  }

  const uint64_t cnt = (uint64_t)m * (uint64_t)n;
  int64_t avgPx[3] = {
      sum[0] / cnt,
      sum[1] / cnt,
      sum[2] / cnt,
  };

  sum[0] = sum[1] = sum[2] = 0;
  for (int y = 0; y < n; y++) {
    const int ys = y * s;
    for (int x = 0; x < m; x++) {
      const int ix = ys + x * four;

      sum[0] += iabs((int64_t)pix[ix + 0] - avgPx[0]);
      sum[1] += iabs((int64_t)pix[ix + 1] - avgPx[1]);
      sum[2] += iabs((int64_t)pix[ix + 2] - avgPx[2]);
    }
  }

  ret->avg[0] = (float64)avgPx[0] / 255.;
  ret->avg[1] = (float64)avgPx[1] / 255.;
  ret->avg[2] = (float64)avgPx[2] / 255.;

  ret->avgDev[0] = (float64)sum[0] / (float64)cnt / 255.;
  ret->avgDev[1] = (float64)sum[1] / (float64)cnt / 255.;
  ret->avgDev[2] = (float64)sum[2] / (float64)cnt / 255.;
}

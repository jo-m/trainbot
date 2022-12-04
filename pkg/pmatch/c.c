#include "c.h"

#include <math.h>

void SearchGrayC(const int m, const int n, const int du, const int dv,
                 const int is, const int ps,
                 /* pixels */
                 const uint8_t* const imgPix, const uint8_t* const patPix,
                 /* return parameters */
                 int* maxX, int* maxY, float64* maxScore) {
#pragma omp parallel for
  for (int y = 0; y < n; y++) {
#pragma omp parallel for
    for (int x = 0; x < m; x++) {
      const int imgPatStartIx = y * is + x;

      uint64_t dot = 0, sqSumI = 0, sqSumP = 0;

      for (int v = 0; v < dv; v++) {
        int pxIi = v * is;
        int pxPi = v * ps;

        for (int u = 0; u < du; u++) {
          const int pxI = imgPix[imgPatStartIx + pxIi];
          const int pxP = patPix[pxPi];

          dot += (uint64_t)(pxI) * (uint64_t)(pxP);
          sqSumI += (uint64_t)(pxI) * (uint64_t)(pxI);
          sqSumP += (uint64_t)(pxP) * (uint64_t)(pxP);

          pxIi++;
          pxPi++;
        }
      }

      const float64 abs = (float64)(sqSumI) * (float64)(sqSumP);
      float64 score;
      if (abs == 0) {
        score = 1;
      } else {
        score = (float64)(dot * dot) / abs;
      }

      if (score > *maxScore) {
        *maxScore = score;
        *maxX = x;
        *maxY = y;
      }
    }
  }

  // this was left out above
  *maxScore = sqrt(*maxScore);
}

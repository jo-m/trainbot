#include "c.h"

#include <math.h>

void SearchGrayC(const int m, const int n, const int du, const int dv,
                 const int is, const int ps, const int imgX0, const int imgY0,
                 const int patX0, const int patY0,
                 /* pixels */
                 const uint8_t* const imgPix, const uint8_t* const patPix,
                 /* return parameters */
                 int* maxX, int* maxY, float64* maxScore) {
  for (int y = 0; y < n; y++) {
    for (int x = 0; x < m; x++) {
      const int winX0 = imgX0 + x;
      const int winY0 = imgY0 + y;
      const int imgPatStartIx = y * is + x;

      uint64_t dot = 0, sqSumI = 0, sqSumP = 0;

      for (int v = 0; v < dv; v++) {
        for (int u = 0; u < du; u++) {
          const int pxIi =
              ((winY0 + v) - winY0) * is + ((winX0 + u) - winX0) * 1;
          const int pxI = imgPix[imgPatStartIx + pxIi];

          const int pxPi =
              ((patY0 + v) - patY0) * ps + ((patX0 + u) - patX0) * 1;
          const int pxP = patPix[pxPi];

          dot += (uint64_t)(pxI) * (uint64_t)(pxP);
          sqSumI += (uint64_t)(pxI) * (uint64_t)(pxI);
          sqSumP += (uint64_t)(pxP) * (uint64_t)(pxP);
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

#include "c.h"

void SearchGrayC(const int m, const int n, const int du, const int dv,
                 const int is, const int ps,
                 /* pixels */
                 const uint8_t* const imgPix, const uint8_t* const patPix,
                 /* return parameters */
                 int* maxX, int* maxY, float64* maxCos2) {
#ifdef _OPENMP
#pragma omp parallel for collapse(2)
#endif
  for (int y = 0; y < n; y++) {
    for (int x = 0; x < m; x++) {
      const int imgPatStartIx = y * is + x;

      uint64_t dot = 0, sqSumI = 0, sqSumP = 0;

      for (int v = 0; v < dv; v++) {
        int pxIi = v * is;
        int pxPi = v * ps;

        for (int u = 0; u < du; u++) {
          const int pxI = imgPix[imgPatStartIx + pxIi + u];
          const int pxP = patPix[pxPi + u];

          dot += (uint64_t)(pxI) * (uint64_t)(pxP);
          sqSumI += (uint64_t)(pxI) * (uint64_t)(pxI);
          sqSumP += (uint64_t)(pxP) * (uint64_t)(pxP);
        }
      }

      const float64 abs = (float64)(sqSumI) * (float64)(sqSumP);
      float64 cos2;
      if (abs == 0) {
        cos2 = 1;
      } else {
        cos2 = (float64)(dot * dot) / abs;
      }

#ifdef _OPENMP
#pragma omp critical
#endif
      if (cos2 > *maxCos2) {
        *maxCos2 = cos2;
        *maxX = x;
        *maxY = y;
      }
    }
  }
}

const int four = 4;

void SearchGrayRGBAC(const int m, const int n, const int du, const int dv,
                     const int is, const int ps,
                     /* pixels */
                     const uint8_t* const imgPix, const uint8_t* const patPix,
                     /* return parameters */
                     int* maxX, int* maxY, float64* maxCos2) {
#ifdef _OPENMP
#pragma omp parallel for collapse(2)
#endif
  for (int y = 0; y < n; y++) {
    for (int x = 0; x < m; x++) {
      const int imgPatStartIx = y * is + x * four;

      uint64_t dot = 0, sqSumI = 0, sqSumP = 0;

      for (int v = 0; v < dv; v++) {
        int pxIi = v * is;
        int pxPi = v * ps;

        for (int u = 0; u < du; u++) {
          for (int rgb = 0; rgb < 3; rgb++) {
            const int pxI = imgPix[imgPatStartIx + pxIi + u * four + rgb];
            const int pxP = patPix[pxPi + u * four + rgb];

            dot += (uint64_t)(pxI) * (uint64_t)(pxP);
            sqSumI += (uint64_t)(pxI) * (uint64_t)(pxI);
            sqSumP += (uint64_t)(pxP) * (uint64_t)(pxP);
          }
        }
      }

      const float64 abs = (float64)(sqSumI) * (float64)(sqSumP);
      float64 cos2;
      if (abs == 0) {
        cos2 = 1;
      } else {
        cos2 = (float64)(dot * dot) / abs;
      }

#ifdef _OPENMP
#pragma omp critical
#endif
      if (cos2 > *maxCos2) {
        *maxCos2 = cos2;
        *maxX = x;
        *maxY = y;
      }
    }
  }
}

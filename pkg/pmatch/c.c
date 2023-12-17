#include "c.h"

static const int four = 4;

void SearchRGBAC(const int m, const int n, const int du, const int dv,
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

      uint64_t dot = 0, absI2 = 0, absP2 = 0;

      for (int v = 0; v < dv; v++) {
        int pxIi = v * is;
        int pxPi = v * ps;

        for (int u = 0; u < du; u++) {
          for (int rgb = 0; rgb < 3; rgb++) {
            const int pxI = imgPix[imgPatStartIx + pxIi + u * four + rgb];
            const int pxP = patPix[pxPi + u * four + rgb];

            dot += (uint64_t)(pxI) * (uint64_t)(pxP);
            absI2 += (uint64_t)(pxI) * (uint64_t)(pxI);
            absP2 += (uint64_t)(pxP) * (uint64_t)(pxP);
          }
        }
      }

      const float64 abs2 = (float64)(absI2) * (float64)(absP2);
      float64 cos2;
      if (abs2 == 0) {
        cos2 = 1;
      } else {
        cos2 = (float64)dot * (float64)dot / abs2;
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

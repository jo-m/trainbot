#pragma once

#include <stdint.h>

#define float64 double

void SearchGrayC(const int m, const int n, const int du, const int dv,
                 const int is, const int ps, const int imgX0, const int imgY0,
                 const int patX0, const int patY0,
                 /* pixels */
                 const uint8_t* const imgPix, const uint8_t* const patPix,
                 /* return parameters */
                 int* maxX, int* maxY, float64* maxScore);

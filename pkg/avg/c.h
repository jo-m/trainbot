#pragma once

#include <stdint.h>

#define float64 double

typedef struct retData {
  float64 avg[3];
  float64 avgDev[3];
} retData;

void RGBAC(const int m, const int n, const int s,
           /* pixels */
           const uint8_t* const pix,
           /* return parameters */
           retData* ret);

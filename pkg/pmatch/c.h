#pragma once

#include <stdint.h>

#define float64 double

void SearchGrayC(const int m, const int n, const int du, const int dv,
                 const int is, const int ps,
                 /* pixels */
                 const uint8_t* const imgPix, const uint8_t* const patPix,
                 /* return parameters */
                 int* maxX, int* maxY, float64* maxCos2);

void SearchGrayRGBAC(const int m, const int n, const int du, const int dv,
                     const int is, const int ps,
                     /* pixels */
                     const uint8_t* const imgPix, const uint8_t* const patPix,
                     /* return parameters */
                     int* maxX, int* maxY, float64* maxCos2);

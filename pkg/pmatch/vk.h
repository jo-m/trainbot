#pragma once

#include <stddef.h>
#include <stdint.h>

// TODO: split generic vk functionality into separate files.
// TODO: support specialization constants (VkSpecializationInfo)
// TODO: support push constants
// TODO: proper error handling, get rid of asserts
// TODO: local size, workgroup size

// Buffers: uint8_t *
// Sizes: size_t
// Dimensions, counts: uint32_t

typedef struct dimensions {
  // Search rect size.
  uint32_t m;
  uint32_t n;
  // Pattern size.
  uint32_t du;
  uint32_t dv;
  // Strides.
  uint32_t is;  // Img.
  uint32_t ps;  // Pat.
  uint32_t ss;  // Search.
  // Output.
  uint32_t max_uint;
  float max;
  uint32_t max_x;
  uint32_t max_y;
} dimensions;

typedef struct dim3 {
  uint32_t x;
  uint32_t y;
  uint32_t z;
} dim3;

void prepare(size_t img_sz, size_t pat_sz, size_t search_sz,
             const uint8_t *shader, size_t shader_sz);
void run(dimensions *dims, const uint8_t *img, const uint8_t *pat,
         uint8_t *search);
void cleanup();

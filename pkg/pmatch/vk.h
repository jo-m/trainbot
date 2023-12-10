#pragma once

#include <stddef.h>
#include <stdint.h>

// TODO: compiler flags
// TODO: split generic vk functionality into separate files.

// Buffers: uint8_t *
// Sizes: size_t
// Dimensions, counts: uint32_t

typedef struct results {
  uint32_t max_uint;
  float max;
  uint32_t max_x;
  uint32_t max_y;
} results;

typedef struct dim3 {
  uint32_t x;
  uint32_t y;
  uint32_t z;
} dim3;

// TODO: replace with VkResult.
typedef int vk_result;
#define vk_success (0)

vk_result prepare(size_t img_sz, size_t pat_sz, size_t search_sz,
                  const uint8_t *shader, size_t shader_sz,
                  int32_t *spec_consts, uint32_t spec_consts_count)
    __attribute((warn_unused_result));
vk_result run(results *out, const uint8_t *img, const uint8_t *pat,
              uint8_t *search, const dim3 wg_sz)
    __attribute((warn_unused_result));
void cleanup();

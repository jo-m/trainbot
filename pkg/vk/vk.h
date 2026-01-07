#pragma once

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <vulkan/vulkan.h>

// Buffers: uint8_t *
// Sizes: size_t
// Dimensions, counts: uint32_t

typedef struct dim3 {
  uint32_t x;
  uint32_t y;
  uint32_t z;
} dim3;

typedef struct vk_handle {
  VkInstance instance;
  VkPhysicalDevice physical_device;
  VkDevice device;
  VkQueue queue;
  uint32_t queue_family_index;
} vk_handle;

VkResult check_big_endian();

VkResult create_vk_handle(vk_handle* out, const bool enable_validation);

void vk_handle_destroy(vk_handle* vk);

int vk_handle_get_device_string(vk_handle* vk, char* str, size_t str_sz);

// Buffer which can be used as VK_DESCRIPTOR_TYPE_STORAGE_BUFFER.
typedef struct vk_buffer {
  VkBuffer buffer;
  VkDeviceMemory buffer_memory;
  VkDeviceSize sz;
} vk_buffer;

// Creates a vk_buffer.
// Usage is hardcoded to VK_BUFFER_USAGE_STORAGE_BUFFER_BIT.
VkResult create_vk_buffer(vk_handle* handle, vk_buffer* out, const size_t sz);

VkResult vk_buffer_read(vk_handle* handle, void* dst, const vk_buffer* src,
                        const size_t sz);

VkResult vk_buffer_write(vk_handle* handle, vk_buffer* dst, const void* src,
                         const size_t sz);

VkResult vk_buffer_zero(vk_handle* handle, vk_buffer* dst, const size_t sz);

void vk_buffer_destroy(vk_handle* handle, vk_buffer* buf);

typedef struct vk_descriptors {
  VkDescriptorPool pool;
  VkDescriptorSetLayout layout;
  VkDescriptorSetLayoutBinding* bindings;
  uint32_t count;
  VkDescriptorSet set;
} vk_descriptors;

typedef struct vk_pipe {
  vk_descriptors desc;

  VkPipelineLayout layout;
  VkPipeline pipeline;

  VkPushConstantRange push_constants;

  VkCommandPool command_pool;
  VkCommandBuffer command_buffer;
} vk_pipe;

VkSpecializationInfo* alloc_int32_spec_info(int32_t* data, uint32_t count);

VkResult create_vk_pipe(vk_handle* handle, vk_pipe* out,
                        const uint8_t* shader_code,
                        const size_t shader_code_sz, const vk_buffer* buffers,
                        const VkDescriptorType* descriptor_types,
                        const uint32_t descriptor_types_count,
                        const VkSpecializationInfo spec_info,
                        const VkPushConstantRange push_constants);

VkResult vk_pipe_run(vk_handle* handle, vk_pipe* pipe, const dim3 wg_sz,
                     const uint8_t* push_constants,
                     const size_t push_constants_sz);

void vk_pipe_destroy(vk_handle* handle, vk_pipe* pipe);

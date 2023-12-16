#include "vk.h"

#include <arpa/inet.h>
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifndef VK_VERSION_1_3
#error Need at least Vulkan SDK 1.3
#endif

#define RET_ON_ERR(expr)             \
  {                                  \
    const VkResult _retval = (expr); \
    if (_retval != VK_SUCCESS) {     \
      return _retval;                \
    }                                \
  }

#define RET_ON_ERR_CLEANUP(expr, cleanup) \
  {                                       \
    const VkResult _retval = (expr);      \
    if (_retval != VK_SUCCESS) {          \
      cleanup;                            \
      return _retval;                     \
    }                                     \
  }

VkResult check_big_endian() {
  if (ntohl(0x01020304) == 0x04030201) {
    return VK_SUCCESS;
  }
  return VK_ERROR_UNKNOWN;
}

static int32_t select_device(const VkPhysicalDevice* devices, uint32_t count) {
  if (count == 0) {
    return -1;
  }

  for (uint32_t i = 0; i < count; i++) {
    VkPhysicalDeviceProperties props = {0};
    vkGetPhysicalDeviceProperties(devices[i], &props);

    // We'll take anything that is a real GPU.
    if (props.deviceType != VK_PHYSICAL_DEVICE_TYPE_CPU &&
        props.deviceType != VK_PHYSICAL_DEVICE_TYPE_OTHER) {
      return i;
    }
  }

  // Otherwise, default to the first one.
  return 0;
}

// Returns the index of a queue family that supports compute operations.
static uint32_t get_compute_queue_family_index(
    const VkPhysicalDevice physical_device) {
  // Query queue families.
  uint32_t queue_family_count;
  vkGetPhysicalDeviceQueueFamilyProperties(physical_device,
                                           &queue_family_count, NULL);
  VkQueueFamilyProperties* const queue_families =
      (VkQueueFamilyProperties*)malloc(sizeof(VkQueueFamilyProperties) *
                                       queue_family_count);
  memset(queue_families, 0,
         sizeof(VkQueueFamilyProperties) * queue_family_count);
  vkGetPhysicalDeviceQueueFamilyProperties(
      physical_device, &queue_family_count, queue_families);

  uint32_t i = 0;
  for (; i < queue_family_count; ++i) {
    VkQueueFamilyProperties props = queue_families[i];

    // Supports compute.
    if (props.queueCount > 0 && (props.queueFlags & VK_QUEUE_COMPUTE_BIT)) {
      break;
    }
  }
  // No queue family supporting compute? This should not happen.
  assert(i < queue_family_count);

  free(queue_families);

  return i;
}

static VkResult create_device(vk_handle* handle)
    __attribute((warn_unused_result));
static VkResult create_device(vk_handle* handle) {
  // Specify queue(s).
  VkDeviceQueueCreateInfo queue_create_info = {0};
  queue_create_info.sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO;
  queue_create_info.flags = 0;
  handle->queue_family_index =
      get_compute_queue_family_index(handle->physical_device);
  queue_create_info.queueFamilyIndex = handle->queue_family_index;
  queue_create_info.queueCount = 1;
  float queue_priorities = 1.0;
  queue_create_info.pQueuePriorities = &queue_priorities;

  // Create logical device.
  VkDeviceCreateInfo device_create_info = {0};
  device_create_info.sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO;
  device_create_info.enabledLayerCount = 0;
  device_create_info.ppEnabledLayerNames = NULL;
  device_create_info.pQueueCreateInfos = &queue_create_info;
  device_create_info.queueCreateInfoCount = 1;
  // No specific features required.
  VkPhysicalDeviceFeatures device_features = {0};
  device_create_info.pEnabledFeatures = &device_features;
  RET_ON_ERR(vkCreateDevice(handle->physical_device, &device_create_info, NULL,
                            &handle->device));

  // Get handle to the queue.
  vkGetDeviceQueue(handle->device, handle->queue_family_index, 0,
                   &handle->queue);

  return VK_SUCCESS;
}

const char* validation_layer = "VK_LAYER_KHRONOS_validation";

static VkResult check_have_validation_layer()
    __attribute((warn_unused_result));
static VkResult check_have_validation_layer() {
  uint32_t layer_count;
  RET_ON_ERR(vkEnumerateInstanceLayerProperties(&layer_count, NULL));
  VkLayerProperties* layer_properties =
      (VkLayerProperties*)malloc(sizeof(VkLayerProperties) * layer_count);
  RET_ON_ERR_CLEANUP(
      vkEnumerateInstanceLayerProperties(&layer_count, layer_properties),
      free(layer_properties));

  bool validation_layer_found = false;
  for (uint32_t i = 0; i < layer_count; i++) {
    VkLayerProperties p = layer_properties[i];
    if (strcmp(validation_layer, p.layerName)) {
      validation_layer_found = true;
      break;
    }
  }

  free(layer_properties);

  if (validation_layer_found) {
    return VK_SUCCESS;
  }
  return VK_ERROR_UNKNOWN;
}

static VkResult check_extensions(const char** names, const uint32_t count)
    __attribute((warn_unused_result));
static VkResult check_extensions(const char** names, const uint32_t count) {
  uint32_t extension_count;
  RET_ON_ERR(
      vkEnumerateInstanceExtensionProperties(NULL, &extension_count, NULL));
  VkExtensionProperties* extension_properties = (VkExtensionProperties*)malloc(
      sizeof(VkExtensionProperties) * extension_count);
  RET_ON_ERR_CLEANUP(vkEnumerateInstanceExtensionProperties(
                         NULL, &extension_count, extension_properties),
                     free(extension_properties));

  // Quadratic... whatever.
  uint32_t found = 0;
  for (uint32_t i = 0; i < extension_count; i++) {
    for (uint32_t j = 0; j < count; j++) {
      if (strcmp(extension_properties[i].extensionName, names[j]) == 0) {
        found++;
      }
    }
  }

  free(extension_properties);
  if (count == found) {
    return VK_SUCCESS;
  }
  return VK_ERROR_UNKNOWN;
}

VkResult create_vk_handle(vk_handle* out, const bool enable_validation) {
  VkApplicationInfo info = {0};
  info.sType = VK_STRUCTURE_TYPE_APPLICATION_INFO;
  info.pNext = NULL;
  info.pApplicationName = "vk.c";
  info.applicationVersion = 0;
  info.pEngineName = "";
  info.engineVersion = 0;
  info.apiVersion = VK_API_VERSION_1_0;

  VkInstanceCreateInfo create_info = {0};
  create_info.sType = VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO;
  create_info.pNext = NULL;
  create_info.flags = 0;
  create_info.pApplicationInfo = &info;

  // Layers.
  create_info.enabledLayerCount = 0;
  create_info.ppEnabledLayerNames = NULL;
  if (enable_validation) {
    RET_ON_ERR(check_have_validation_layer());
    create_info.enabledLayerCount = 1;
    create_info.ppEnabledLayerNames = &validation_layer;
  }

  // Extensions.
  const char* extensions[] = {
      VK_EXT_DEBUG_REPORT_EXTENSION_NAME,
  };
  create_info.enabledExtensionCount = 0;
  create_info.ppEnabledExtensionNames = extensions;
  if (enable_validation) {
    RET_ON_ERR(check_extensions(&extensions[0], 1));
    create_info.enabledExtensionCount += 1;
  }

  // Create instance.
  RET_ON_ERR(vkCreateInstance(&create_info, NULL, &out->instance));

  // Query physical devices.
  uint32_t physical_device_count = 0;
  vkEnumeratePhysicalDevices(out->instance, &physical_device_count, NULL);
  VkPhysicalDevice* const physical_devices = (VkPhysicalDevice*)malloc(
      sizeof(VkPhysicalDevice) * physical_device_count);
  memset(physical_devices, 0,
         sizeof(VkPhysicalDevice) * physical_device_count);
  RET_ON_ERR_CLEANUP(
      vkEnumeratePhysicalDevices(out->instance, &physical_device_count,
                                 physical_devices),
      free(physical_devices));

  // Select one.
  const int32_t device_ix =
      select_device(physical_devices, physical_device_count);
  // No device found?
  if (device_ix < 0) {
    free(physical_devices);
    return VK_ERROR_UNKNOWN;
  }
  out->physical_device = physical_devices[device_ix];

  // Create and cleanup.
  RET_ON_ERR_CLEANUP(create_device(out), free(physical_devices));
  free(physical_devices);

  return VK_SUCCESS;
}

int vk_handle_get_device_string(vk_handle* vk, char* str, size_t str_sz) {
  VkPhysicalDeviceProperties props = {0};
  vkGetPhysicalDeviceProperties(vk->physical_device, &props);
  const uint32_t v = props.apiVersion;
  return snprintf(str, str_sz,
                  "name='%s' vendor_id=%u device_id=%u driver_version=%u "
                  "variant=%u api_version=%u.%u.%u type=%u",
                  props.deviceName, props.vendorID, props.deviceID,
                  props.driverVersion, VK_API_VERSION_VARIANT(v),
                  VK_API_VERSION_MAJOR(v), VK_API_VERSION_MINOR(v),
                  VK_API_VERSION_PATCH(v), props.deviceType);
}

void vk_handle_destroy(vk_handle* vk) {
  vkDestroyDevice(vk->device, NULL);
  vkDestroyInstance(vk->instance, NULL);
}

static int32_t find_memory_type(const VkPhysicalDevice phys_device,
                                const uint32_t bits,
                                const VkMemoryPropertyFlags prop_flags) {
  VkPhysicalDeviceMemoryProperties props;
  vkGetPhysicalDeviceMemoryProperties(phys_device, &props);

  for (uint32_t i = 0; i < props.memoryTypeCount; ++i) {
    if ((bits & (1 << i)) &&
        ((props.memoryTypes[i].propertyFlags & prop_flags) == prop_flags))
      return i;
  }
  return -1;
}

VkResult create_vk_buffer(vk_handle* handle, vk_buffer* out, const size_t sz) {
  out->sz = sz;

  // Create buffer.
  VkBufferCreateInfo buffer_create_info = {0};
  buffer_create_info.sType = VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO;
  buffer_create_info.size = out->sz;
  buffer_create_info.usage = VK_BUFFER_USAGE_STORAGE_BUFFER_BIT;
  buffer_create_info.sharingMode = VK_SHARING_MODE_EXCLUSIVE;

  RET_ON_ERR(
      vkCreateBuffer(handle->device, &buffer_create_info, NULL, &out->buffer));

  // Gather memory requirements from device.
  VkMemoryRequirements memory_requirements;
  vkGetBufferMemoryRequirements(handle->device, out->buffer,
                                &memory_requirements);

  VkMemoryAllocateInfo allocate_info = {0};
  allocate_info.sType = VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO;
  allocate_info.allocationSize = memory_requirements.size;

  allocate_info.memoryTypeIndex = find_memory_type(
      handle->physical_device, memory_requirements.memoryTypeBits,
      VK_MEMORY_PROPERTY_HOST_COHERENT_BIT |
          VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT);

  // Allocate.
  RET_ON_ERR(vkAllocateMemory(handle->device, &allocate_info, NULL,
                              &out->buffer_memory));

  // Bind to buffer.
  return vkBindBufferMemory(handle->device, out->buffer, out->buffer_memory,
                            0);
}

VkResult vk_buffer_read(vk_handle* handle, void* dst, const vk_buffer* src,
                        const size_t sz) {
  if (sz != src->sz) {  // TODO: maybe relax
    return VK_ERROR_UNKNOWN;
  }

  void* mapped = NULL;
  RET_ON_ERR(
      vkMapMemory(handle->device, src->buffer_memory, 0, sz, 0, &mapped));
  memcpy(dst, mapped, sz);
  vkUnmapMemory(handle->device, src->buffer_memory);
  return VK_SUCCESS;
}

VkResult vk_buffer_write(vk_handle* handle, vk_buffer* dst, const void* src,
                         const size_t sz) {
  if (sz != dst->sz) {  // TODO: maybe relax
    return VK_ERROR_UNKNOWN;
  }

  void* mapped = NULL;
  RET_ON_ERR(
      vkMapMemory(handle->device, dst->buffer_memory, 0, sz, 0, &mapped));
  memcpy(mapped, src, sz);
  vkUnmapMemory(handle->device, dst->buffer_memory);
  return VK_SUCCESS;
}

void vk_buffer_destroy(vk_handle* handle, vk_buffer* buf) {
  vkFreeMemory(handle->device, buf->buffer_memory, NULL);
  vkDestroyBuffer(handle->device, buf->buffer, NULL);
}

static void vk_descriptors_destroy(vk_handle* handle, vk_descriptors* desc) {
  vkDestroyDescriptorSetLayout(handle->device, desc->layout, NULL);
  vkDestroyDescriptorPool(handle->device, desc->pool, NULL);
  free(desc->bindings);
  desc->bindings = NULL;
  desc->count = 0;
}

static VkResult create_vk_descriptors(vk_descriptors* out, vk_handle* handle,
                                      const VkDescriptorType* descriptor_types,
                                      const uint32_t count)
    __attribute((warn_unused_result));
static VkResult create_vk_descriptors(vk_descriptors* out, vk_handle* handle,
                                      const VkDescriptorType* descriptor_types,
                                      const uint32_t count) {
  out->count = count;

  for (uint32_t i = 0; i < count; i++) {
    // Currently, only VK_DESCRIPTOR_TYPE_STORAGE_BUFFER is supported.
    assert(descriptor_types[i] == VK_DESCRIPTOR_TYPE_STORAGE_BUFFER);
  }
  VkDescriptorPoolSize pool_size = {VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, count};

  // Create descriptor pool.
  VkDescriptorPoolCreateInfo descriptor_pool_create_info = {0};
  descriptor_pool_create_info.sType =
      VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO;
  descriptor_pool_create_info.pNext = NULL;
  descriptor_pool_create_info.flags = 0;
  descriptor_pool_create_info.maxSets = 1;
  descriptor_pool_create_info.poolSizeCount = 1;
  descriptor_pool_create_info.pPoolSizes = &pool_size;

  RET_ON_ERR(vkCreateDescriptorPool(
      handle->device, &descriptor_pool_create_info, NULL, &out->pool));

  // Create descriptor set layout bindings.
  out->bindings = (VkDescriptorSetLayoutBinding*)malloc(
      sizeof(VkDescriptorSetLayoutBinding) * count);
  memset(out->bindings, 0, sizeof(VkDescriptorSetLayoutBinding) * count);

  for (uint32_t i = 0; i < count; i++) {
    VkDescriptorSetLayoutBinding* b = &out->bindings[i];
    b->binding = i;
    b->descriptorType = descriptor_types[i];
    b->descriptorCount = 1;
    b->stageFlags = VK_SHADER_STAGE_COMPUTE_BIT;
    b->pImmutableSamplers = 0;
  }

  // Create descriptor set layout.
  VkDescriptorSetLayoutCreateInfo descriptor_set_layout_create_info = {0};
  descriptor_set_layout_create_info.sType =
      VK_STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO;
  descriptor_set_layout_create_info.pNext = NULL;
  descriptor_set_layout_create_info.flags = 0;
  descriptor_set_layout_create_info.bindingCount = count;
  descriptor_set_layout_create_info.pBindings = out->bindings;

  RET_ON_ERR_CLEANUP(
      vkCreateDescriptorSetLayout(
          handle->device, &descriptor_set_layout_create_info, 0, &out->layout),
      free(out->bindings));

  // Allocate descriptor sets.
  VkDescriptorSetAllocateInfo descriptor_set_allocate_info = {0};
  descriptor_set_allocate_info.sType =
      VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO;
  descriptor_set_allocate_info.pNext = NULL;
  descriptor_set_allocate_info.descriptorPool = out->pool;
  descriptor_set_allocate_info.descriptorSetCount = 1;
  descriptor_set_allocate_info.pSetLayouts = &out->layout;

  RET_ON_ERR_CLEANUP(
      vkAllocateDescriptorSets(handle->device, &descriptor_set_allocate_info,
                               &out->set),
      free(out->bindings));

  return VK_SUCCESS;
}

static void vk_descriptors_bind(vk_handle* handle, vk_descriptors* desc,
                                const vk_buffer* buffers,
                                const VkDescriptorType* descriptor_types,
                                const uint32_t count) {
  assert(count == desc->count);

  for (uint32_t i = 0; i < count; i++) {
    VkDescriptorBufferInfo descriptor_buffer_info = {0};
    descriptor_buffer_info.buffer = buffers[i].buffer;
    descriptor_buffer_info.offset = 0;
    descriptor_buffer_info.range = buffers[i].sz;

    VkWriteDescriptorSet write_descriptor_set = {0};
    write_descriptor_set.sType = VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET;
    write_descriptor_set.pNext = NULL;
    write_descriptor_set.dstSet = desc->set;
    write_descriptor_set.dstBinding = i;
    write_descriptor_set.dstArrayElement = 0;
    write_descriptor_set.descriptorCount = 1;
    write_descriptor_set.descriptorType = descriptor_types[i];
    write_descriptor_set.pBufferInfo = &descriptor_buffer_info;
    write_descriptor_set.pTexelBufferView = NULL;

    vkUpdateDescriptorSets(handle->device, 1, &write_descriptor_set, 0, NULL);
  }
}

VkResult create_vk_pipe(vk_handle* handle, vk_pipe* out,
                        const uint8_t* shader_code,
                        const size_t shader_code_sz, const vk_buffer* buffers,
                        const VkDescriptorType* descriptor_types,
                        const uint32_t descriptor_types_count,
                        const VkSpecializationInfo spec_info,
                        const VkPushConstantRange push_constants) {
  RET_ON_ERR(create_vk_descriptors(&out->desc, handle, descriptor_types,
                                   descriptor_types_count));
  vk_descriptors_bind(handle, &out->desc, buffers, descriptor_types,
                      descriptor_types_count);

  // Create shader.
  VkShaderModule shader;
  if (shader_code_sz % 4 != 0) {
    return VK_SUCCESS;
  }
  VkShaderModuleCreateInfo create_info = {0};
  create_info.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO;
  create_info.pNext = NULL;
  create_info.flags = 0;
  create_info.pCode = (uint32_t*)shader_code;
  create_info.codeSize = shader_code_sz;

  RET_ON_ERR(
      vkCreateShaderModule(handle->device, &create_info, NULL, &shader));

  // Create layout.
  VkPipelineLayoutCreateInfo pipeline_layout_create_info = {0};
  pipeline_layout_create_info.sType =
      VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO;
  pipeline_layout_create_info.pNext = NULL;
  pipeline_layout_create_info.flags = 0;
  pipeline_layout_create_info.setLayoutCount = 1;
  pipeline_layout_create_info.pSetLayouts = &out->desc.layout;
  pipeline_layout_create_info.pushConstantRangeCount =
      push_constants.size > 0 ? 1 : 0;
  pipeline_layout_create_info.pPushConstantRanges = &push_constants;
  RET_ON_ERR(vkCreatePipelineLayout(
      handle->device, &pipeline_layout_create_info, NULL, &out->layout));
  out->push_constants = push_constants;

  // Create pipeline.
  VkPipelineShaderStageCreateInfo shader_stage_create_info = {0};
  shader_stage_create_info.sType =
      VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
  shader_stage_create_info.pNext = NULL;
  shader_stage_create_info.flags = 0;
  shader_stage_create_info.stage = VK_SHADER_STAGE_COMPUTE_BIT;
  shader_stage_create_info.module = shader;
  shader_stage_create_info.pName = "main";
  shader_stage_create_info.pSpecializationInfo = &spec_info;

  VkComputePipelineCreateInfo pipeline_create_info = {0};
  pipeline_create_info.sType = VK_STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO;
  pipeline_create_info.pNext = NULL;
  pipeline_create_info.flags = 0;
  pipeline_create_info.stage = shader_stage_create_info;
  pipeline_create_info.layout = out->layout;

  RET_ON_ERR(vkCreateComputePipelines(handle->device, VK_NULL_HANDLE, 1,
                                      &pipeline_create_info, NULL,
                                      &out->pipeline));

  vkDestroyShaderModule(handle->device, shader, NULL);

  // Create command pool.
  VkCommandPoolCreateInfo command_pool_create_info = {0};
  command_pool_create_info.sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO;
  command_pool_create_info.pNext = NULL;
  command_pool_create_info.flags =
      VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT;
  command_pool_create_info.queueFamilyIndex = handle->queue_family_index;
  RET_ON_ERR(vkCreateCommandPool(handle->device, &command_pool_create_info,
                                 NULL, &out->command_pool));

  // Allocate command buffer.
  VkCommandBufferAllocateInfo command_buffer_allocate_info = {0};
  command_buffer_allocate_info.sType =
      VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO;
  command_buffer_allocate_info.pNext = NULL;
  command_buffer_allocate_info.commandPool = out->command_pool;
  command_buffer_allocate_info.level = VK_COMMAND_BUFFER_LEVEL_PRIMARY;
  command_buffer_allocate_info.commandBufferCount = 1;
  return vkAllocateCommandBuffers(
      handle->device, &command_buffer_allocate_info, &out->command_buffer);
}

VkResult vk_pipe_run(vk_handle* handle, vk_pipe* pipe, const dim3 wg_sz,
                     const uint8_t* push_constants,
                     const size_t push_constants_sz) {
  // Begin command buffer.
  VkCommandBufferBeginInfo begin_info = {0};
  begin_info.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO;
  begin_info.pNext = NULL;
  begin_info.pInheritanceInfo = NULL;
  begin_info.flags = VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT;
  RET_ON_ERR(vkBeginCommandBuffer(pipe->command_buffer, &begin_info));

  // Push constants.
  if (push_constants != NULL && push_constants_sz > 0) {
    assert(push_constants_sz <= pipe->push_constants.size);
    vkCmdPushConstants(pipe->command_buffer, pipe->layout,
                       VK_SHADER_STAGE_COMPUTE_BIT, 0, push_constants_sz,
                       push_constants);
  }

  // Bind pipeline and descriptor set.
  vkCmdBindPipeline(pipe->command_buffer, VK_PIPELINE_BIND_POINT_COMPUTE,
                    pipe->pipeline);
  vkCmdBindDescriptorSets(pipe->command_buffer, VK_PIPELINE_BIND_POINT_COMPUTE,
                          pipe->layout, 0, 1, &pipe->desc.set, 0, NULL);

  // Dispatch.
  vkCmdDispatch(pipe->command_buffer, wg_sz.x, wg_sz.y, wg_sz.z);
  RET_ON_ERR(vkEndCommandBuffer(pipe->command_buffer));

  // Create fence.
  VkFence fence;
  VkFenceCreateInfo fence_create_info = {0};
  fence_create_info.sType = VK_STRUCTURE_TYPE_FENCE_CREATE_INFO;
  fence_create_info.pNext = NULL;
  fence_create_info.flags = 0;
  RET_ON_ERR(vkCreateFence(handle->device, &fence_create_info, NULL, &fence));

  // Submit to queue and wait.
  VkSubmitInfo submit_info = {0};
  submit_info.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO;
  submit_info.pNext = NULL;
  submit_info.waitSemaphoreCount = 0;
  submit_info.pWaitSemaphores = NULL;
  submit_info.pWaitDstStageMask = NULL;
  submit_info.commandBufferCount = 1;
  submit_info.pCommandBuffers = &pipe->command_buffer;
  submit_info.signalSemaphoreCount = 0;
  submit_info.pSignalSemaphores = NULL;

  RET_ON_ERR(vkQueueSubmit(handle->queue, 1, &submit_info, fence));
  RET_ON_ERR(
      vkWaitForFences(handle->device, 1, &fence, VK_TRUE, 100000000000));

  // Cleanup.
  vkDestroyFence(handle->device, fence, NULL);
  return VK_SUCCESS;
}

void vk_pipe_destroy(vk_handle* handle, vk_pipe* pipe) {
  vkFreeCommandBuffers(handle->device, pipe->command_pool, 1,
                       &pipe->command_buffer);
  vkDestroyCommandPool(handle->device, pipe->command_pool, NULL);
  vkDestroyPipelineLayout(handle->device, pipe->layout, NULL);
  vkDestroyPipeline(handle->device, pipe->pipeline, NULL);

  vk_descriptors_destroy(handle, &pipe->desc);
}

typedef struct vk_spec_info {
  VkSpecializationInfo info;
  VkSpecializationMapEntry map[];
} vk_spec_info;

// Caller must free retval after use.
// Retval will point to *data.
VkSpecializationInfo* alloc_int32_spec_info(int32_t* data, uint32_t count) {
  vk_spec_info* ret =
      malloc(sizeof(vk_spec_info) + count * sizeof(VkSpecializationMapEntry));

  for (uint32_t i = 0; i < count; i++) {
    ret->map[i].constantID = i;
    ret->map[i].offset = ((uint8_t*)&data[i]) - ((uint8_t*)&data[0]);
    ret->map[i].size = sizeof(*data);
  }

  ret->info.mapEntryCount = count;
  ret->info.pMapEntries = &ret->map[0];
  ret->info.dataSize = sizeof(*data) * count;
  ret->info.pData = data;

  return (VkSpecializationInfo*)ret;
}

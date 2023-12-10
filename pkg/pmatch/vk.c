#include "vk.h"

#include <arpa/inet.h>
#include <assert.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <vulkan/vulkan.h>

#ifndef VK_VERSION_1_3
#error Need at least Vulkan SDK 1.3
#endif

#define EXIT_ON_VULKAN_FAILURE(result)                         \
  if (VK_SUCCESS != (result)) {                                \
    fprintf(stderr, "Failure at %u %s\n", __LINE__, __FILE__); \
    exit(-1);                                                  \
  }

typedef struct vk_handle {
  VkInstance instance;
  VkPhysicalDevice physical_device;
  VkDevice device;
  VkQueue queue;
  uint32_t queue_family_index;
} vk_handle;

static int32_t select_device(const VkPhysicalDevice* devices, uint32_t count) {
  assert(count > 0);
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

  assert(i < queue_family_count);

  free(queue_families);

  return i;
}

static void create_device(vk_handle* handle) {
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
  EXIT_ON_VULKAN_FAILURE(vkCreateDevice(
      handle->physical_device, &device_create_info, NULL, &handle->device));

  // Get handle to the queue.
  vkGetDeviceQueue(handle->device, handle->queue_family_index, 0,
                   &handle->queue);
}

const char* validation_layer = "VK_LAYER_KHRONOS_validation";

static void check_have_validation_layer() {
  uint32_t layer_count;
  vkEnumerateInstanceLayerProperties(&layer_count, NULL);
  VkLayerProperties* layer_properties =
      (VkLayerProperties*)malloc(sizeof(VkLayerProperties) * layer_count);
  vkEnumerateInstanceLayerProperties(&layer_count, layer_properties);

  bool validation_layer_found = false;
  for (uint32_t i = 0; i < layer_count; i++) {
    VkLayerProperties p = layer_properties[i];
    if (strcmp(validation_layer, p.layerName)) {
      validation_layer_found = true;
      break;
    }
  }

  free(layer_properties);
  assert(validation_layer_found);
}

static void check_extensions(const char** names, const uint32_t count) {
  uint32_t extension_count;
  vkEnumerateInstanceExtensionProperties(NULL, &extension_count, NULL);
  VkExtensionProperties* extension_properties = (VkExtensionProperties*)malloc(
      sizeof(VkExtensionProperties) * extension_count);
  vkEnumerateInstanceExtensionProperties(NULL, &extension_count,
                                         extension_properties);

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
  assert(count == found);
}

static vk_handle create_vk_handle(const bool enable_validation) {
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
    check_have_validation_layer();
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
    check_extensions(&extensions[0], 1);
    create_info.enabledExtensionCount += 1;
  }

  // Create instance.
  vk_handle handle = {0};
  EXIT_ON_VULKAN_FAILURE(
      vkCreateInstance(&create_info, NULL, &handle.instance));

  // Query physical devices.
  uint32_t physical_device_count = 0;
  vkEnumeratePhysicalDevices(handle.instance, &physical_device_count, NULL);
  VkPhysicalDevice* const physical_devices = (VkPhysicalDevice*)malloc(
      sizeof(VkPhysicalDevice) * physical_device_count);
  memset(physical_devices, 0,
         sizeof(VkPhysicalDevice) * physical_device_count);
  EXIT_ON_VULKAN_FAILURE(vkEnumeratePhysicalDevices(
      handle.instance, &physical_device_count, physical_devices));

  // Select one.
  const uint32_t device_ix =
      select_device(physical_devices, physical_device_count);
  assert(device_ix >= 0);
  handle.physical_device = physical_devices[device_ix];

  // Create and cleanup.
  create_device(&handle);
  free(physical_devices);

  return handle;
}

static int vk_handle_get_device_string(vk_handle* vk, char* str,
                                       size_t str_sz) {
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

static void vk_handle_destroy(vk_handle* vk) {
  vkDestroyDevice(vk->device, NULL);
  vkDestroyInstance(vk->instance, NULL);
}

int32_t find_memory_type(const VkPhysicalDevice phys_device,
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

// Buffer which can be used as VK_DESCRIPTOR_TYPE_STORAGE_BUFFER.
typedef struct vk_buffer {
  VkBuffer buffer;
  VkDeviceMemory buffer_memory;
  VkDeviceSize sz;
} vk_buffer;

// Creates a vk_buffer.
// Usage is hardcoded to VK_BUFFER_USAGE_STORAGE_BUFFER_BIT.
static vk_buffer create_vk_buffer(vk_handle* handle, const size_t sz) {
  vk_buffer buf = {0};
  buf.sz = sz;

  // Create buffer.
  VkBufferCreateInfo buffer_create_info = {0};
  buffer_create_info.sType = VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO;
  buffer_create_info.size = buf.sz;
  buffer_create_info.usage = VK_BUFFER_USAGE_STORAGE_BUFFER_BIT;
  buffer_create_info.sharingMode = VK_SHARING_MODE_EXCLUSIVE;

  EXIT_ON_VULKAN_FAILURE(
      vkCreateBuffer(handle->device, &buffer_create_info, NULL, &buf.buffer));

  // Gather memory requirements from device.
  VkMemoryRequirements memory_requirements;
  vkGetBufferMemoryRequirements(handle->device, buf.buffer,
                                &memory_requirements);

  VkMemoryAllocateInfo allocate_info = {0};
  allocate_info.sType = VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO;
  allocate_info.allocationSize = memory_requirements.size;

  allocate_info.memoryTypeIndex = find_memory_type(
      handle->physical_device, memory_requirements.memoryTypeBits,
      VK_MEMORY_PROPERTY_HOST_COHERENT_BIT |
          VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT);

  // Allocate.
  EXIT_ON_VULKAN_FAILURE(vkAllocateMemory(handle->device, &allocate_info, NULL,
                                          &buf.buffer_memory));

  // Bind to buffer.
  EXIT_ON_VULKAN_FAILURE(
      vkBindBufferMemory(handle->device, buf.buffer, buf.buffer_memory, 0));

  return buf;
}

static void vk_buffer_read(vk_handle* handle, const vk_buffer* src, void* dst,
                           const size_t sz) {
  assert(sz == src->sz);
  void* mapped = NULL;
  vkMapMemory(handle->device, src->buffer_memory, 0, sz, 0, &mapped);

  memcpy(dst, mapped, sz);

  vkUnmapMemory(handle->device, src->buffer_memory);
}

static void vk_buffer_write(vk_handle* handle, vk_buffer* dst, const void* src,
                            const size_t sz) {
  assert(sz == dst->sz);
  void* mapped = NULL;
  vkMapMemory(handle->device, dst->buffer_memory, 0, sz, 0, &mapped);

  memcpy(mapped, src, sz);

  vkUnmapMemory(handle->device, dst->buffer_memory);
}

static void vk_buffer_destroy(vk_handle* handle, vk_buffer* buf) {
  vkFreeMemory(handle->device, buf->buffer_memory, NULL);
  vkDestroyBuffer(handle->device, buf->buffer, NULL);
}

typedef struct _vk_descriptors {
  VkDescriptorPool pool;
  VkDescriptorSetLayout layout;
  VkDescriptorSetLayoutBinding* bindings;
  uint32_t count;
  VkDescriptorSet set;
} _vk_descriptors;

static void _vk_descriptors_destroy(vk_handle* handle, _vk_descriptors* desc) {
  vkDestroyDescriptorSetLayout(handle->device, desc->layout, NULL);
  vkDestroyDescriptorPool(handle->device, desc->pool, NULL);
  free(desc->bindings);
  desc->bindings = NULL;
  desc->count = 0;
}

static _vk_descriptors _create_vk_descriptors(
    vk_handle* handle, const VkDescriptorType* descriptor_types,
    const uint32_t count) {
  _vk_descriptors desc = {0};
  desc.count = count;

  // For now, we simply assume all buffers are bound the same way.
  // TODO: merge counts from descriptor_types into different pool sizes.
  for (uint32_t i = 0; i < count; i++) {
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

  EXIT_ON_VULKAN_FAILURE(vkCreateDescriptorPool(
      handle->device, &descriptor_pool_create_info, NULL, &desc.pool));

  // Create descriptor set layout bindings.
  desc.bindings = (VkDescriptorSetLayoutBinding*)malloc(
      sizeof(VkDescriptorSetLayoutBinding) * count);
  memset(desc.bindings, 0, sizeof(VkDescriptorSetLayoutBinding) * count);

  for (uint32_t i = 0; i < count; i++) {
    VkDescriptorSetLayoutBinding* b = &desc.bindings[i];
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
  descriptor_set_layout_create_info.pBindings = desc.bindings;

  EXIT_ON_VULKAN_FAILURE(vkCreateDescriptorSetLayout(
      handle->device, &descriptor_set_layout_create_info, 0, &desc.layout));

  // Allocate descriptor sets.
  VkDescriptorSetAllocateInfo descriptor_set_allocate_info = {0};
  descriptor_set_allocate_info.sType =
      VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO;
  descriptor_set_allocate_info.pNext = NULL;
  descriptor_set_allocate_info.descriptorPool = desc.pool;
  descriptor_set_allocate_info.descriptorSetCount = 1;
  descriptor_set_allocate_info.pSetLayouts = &desc.layout;

  EXIT_ON_VULKAN_FAILURE(vkAllocateDescriptorSets(
      handle->device, &descriptor_set_allocate_info, &desc.set));

  return desc;
}

static void _vk_descriptors_bind(vk_handle* handle, _vk_descriptors* desc,
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

typedef struct vk_pipe {
  _vk_descriptors desc;

  VkPipelineLayout layout;
  VkPipeline pipeline;

  VkCommandPool command_pool;
  VkCommandBuffer command_buffer;
} vk_pipe;

static vk_pipe create_vk_pipe(vk_handle* handle, const uint8_t* shader_code,
                              const size_t shader_code_sz,
                              const vk_buffer* buffers,
                              const VkDescriptorType* descriptor_types,
                              const uint32_t descriptor_types_count,
                              const VkSpecializationInfo spec_info) {
  vk_pipe pipe = {0};

  pipe.desc =
      _create_vk_descriptors(handle, descriptor_types, descriptor_types_count);
  _vk_descriptors_bind(handle, &pipe.desc, buffers, descriptor_types,
                       descriptor_types_count);

  // Create shader.
  VkShaderModule shader;
  assert(shader_code_sz % 4 == 0);
  VkShaderModuleCreateInfo create_info = {0};
  create_info.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO;
  create_info.pNext = NULL;
  create_info.flags = 0;
  create_info.pCode = (uint32_t*)shader_code;
  create_info.codeSize = shader_code_sz;

  EXIT_ON_VULKAN_FAILURE(
      vkCreateShaderModule(handle->device, &create_info, NULL, &shader));

  // Create layout.
  VkPipelineLayoutCreateInfo pipeline_layout_create_info = {0};
  pipeline_layout_create_info.sType =
      VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO;
  pipeline_layout_create_info.pNext = NULL;
  pipeline_layout_create_info.flags = 0;
  pipeline_layout_create_info.setLayoutCount = 1;
  pipeline_layout_create_info.pSetLayouts = &pipe.desc.layout;
  pipeline_layout_create_info.pushConstantRangeCount = 0;
  pipeline_layout_create_info.pPushConstantRanges = NULL;
  EXIT_ON_VULKAN_FAILURE(vkCreatePipelineLayout(
      handle->device, &pipeline_layout_create_info, NULL, &pipe.layout));

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
  pipeline_create_info.layout = pipe.layout;

  EXIT_ON_VULKAN_FAILURE(
      vkCreateComputePipelines(handle->device, VK_NULL_HANDLE, 1,
                               &pipeline_create_info, NULL, &pipe.pipeline));

  vkDestroyShaderModule(handle->device, shader, NULL);

  // Create command pool.
  VkCommandPoolCreateInfo command_pool_create_info = {0};
  command_pool_create_info.sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO;
  command_pool_create_info.pNext = NULL;
  command_pool_create_info.flags =
      VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT;
  command_pool_create_info.queueFamilyIndex = handle->queue_family_index;
  EXIT_ON_VULKAN_FAILURE(vkCreateCommandPool(
      handle->device, &command_pool_create_info, NULL, &pipe.command_pool));

  // Allocate command buffer.
  VkCommandBufferAllocateInfo command_buffer_allocate_info = {0};
  command_buffer_allocate_info.sType =
      VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO;
  command_buffer_allocate_info.pNext = NULL;
  command_buffer_allocate_info.commandPool = pipe.command_pool;
  command_buffer_allocate_info.level = VK_COMMAND_BUFFER_LEVEL_PRIMARY;
  command_buffer_allocate_info.commandBufferCount = 1;
  EXIT_ON_VULKAN_FAILURE(vkAllocateCommandBuffers(
      handle->device, &command_buffer_allocate_info, &pipe.command_buffer));

  return pipe;
}

static void vk_pipe_run(vk_handle* handle, vk_pipe* pipe, const dim3 wg_sz) {
  // Begin command buffer.
  VkCommandBufferBeginInfo begin_info = {0};
  begin_info.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO;
  begin_info.pNext = NULL;
  begin_info.pInheritanceInfo = NULL;
  begin_info.flags = VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT;
  EXIT_ON_VULKAN_FAILURE(
      vkBeginCommandBuffer(pipe->command_buffer, &begin_info));

  // Bind pipeline and descriptor set.
  vkCmdBindPipeline(pipe->command_buffer, VK_PIPELINE_BIND_POINT_COMPUTE,
                    pipe->pipeline);
  vkCmdBindDescriptorSets(pipe->command_buffer, VK_PIPELINE_BIND_POINT_COMPUTE,
                          pipe->layout, 0, 1, &pipe->desc.set, 0, NULL);

  // Dispatch.
  vkCmdDispatch(pipe->command_buffer, wg_sz.x, wg_sz.y, wg_sz.z);
  EXIT_ON_VULKAN_FAILURE(vkEndCommandBuffer(pipe->command_buffer));

  // Create fence.
  VkFence fence;
  VkFenceCreateInfo fence_create_info = {};
  fence_create_info.sType = VK_STRUCTURE_TYPE_FENCE_CREATE_INFO;
  fence_create_info.pNext = NULL;
  fence_create_info.flags = 0;
  EXIT_ON_VULKAN_FAILURE(
      vkCreateFence(handle->device, &fence_create_info, NULL, &fence));

  // Submit to queue and wait.
  VkSubmitInfo submit_info = {};
  submit_info.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO;
  submit_info.pNext = NULL;
  submit_info.waitSemaphoreCount = 0;
  submit_info.pWaitSemaphores = NULL;
  submit_info.pWaitDstStageMask = NULL;
  submit_info.commandBufferCount = 1;
  submit_info.pCommandBuffers = &pipe->command_buffer;
  submit_info.signalSemaphoreCount = 0;
  submit_info.pSignalSemaphores = NULL;

  EXIT_ON_VULKAN_FAILURE(vkQueueSubmit(handle->queue, 1, &submit_info, fence));
  EXIT_ON_VULKAN_FAILURE(
      vkWaitForFences(handle->device, 1, &fence, VK_TRUE, 100000000000));

  // Cleanup.
  vkDestroyFence(handle->device, fence, NULL);
}

static void vk_pipe_destroy(vk_handle* handle, vk_pipe* pipe) {
  vkFreeCommandBuffers(handle->device, pipe->command_pool, 1,
                       &pipe->command_buffer);
  vkDestroyCommandPool(handle->device, pipe->command_pool, NULL);
  vkDestroyPipelineLayout(handle->device, pipe->layout, NULL);
  vkDestroyPipeline(handle->device, pipe->pipeline, NULL);

  _vk_descriptors_destroy(handle, &pipe->desc);
}

void assert_big_endian() { assert(ntohl(0x01020304) == 0x04030201); }

typedef struct vk_spec_info {
  VkSpecializationInfo info;
  VkSpecializationMapEntry map[];
} vk_spec_info;

// Caller must free retval after use.
// Retval will point to *data.
VkSpecializationInfo* alloc_int32_spec_info(int32_t* data, uint32_t count) {
  vk_spec_info* ret =
      malloc(sizeof(vk_spec_info) + count * sizeof(VkSpecializationMapEntry));

  for (int i = 0; i < count; i++) {
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

// TODO: no global state
static vk_handle handle;
static vk_buffer bufs[4];
static vk_pipe pipe;

void prepare(const size_t img_sz, const size_t pat_sz, const size_t search_sz,
             const uint8_t* shader, const size_t shader_sz,
             int32_t* spec_consts, uint32_t spec_consts_count) {
  // Setup.
  handle = create_vk_handle(true);

  char device_str[512] = {0};
  vk_handle_get_device_string(&handle, device_str, sizeof(device_str));
  printf("Selected device: %s\n", device_str);

  bufs[0] = create_vk_buffer(&handle, sizeof(results));
  bufs[1] = create_vk_buffer(&handle, img_sz);
  bufs[2] = create_vk_buffer(&handle, pat_sz);
  bufs[3] = create_vk_buffer(&handle, search_sz);
  VkDescriptorType descriptor_types[4] = {
      VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER,
      VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER};
  assert(sizeof(bufs) / sizeof(bufs[0]) ==
         sizeof(descriptor_types) / sizeof(descriptor_types[0]));

  VkSpecializationInfo* spec_info =
      alloc_int32_spec_info(spec_consts, spec_consts_count);

  pipe = create_vk_pipe(&handle, shader, shader_sz, bufs, descriptor_types,
                        sizeof(descriptor_types) / sizeof(descriptor_types[0]),
                        *spec_info);

  free(spec_info);
}

void run(results* res, const uint8_t* img, const uint8_t* pat, uint8_t* search,
         const dim3 wg_sz) {
  assert_big_endian();

  // Write to input buffers.
  vk_buffer_write(&handle, &bufs[0], res, sizeof(results));
  vk_buffer_write(&handle, &bufs[1], img, bufs[1].sz);
  vk_buffer_write(&handle, &bufs[2], pat, bufs[2].sz);

  // Do some actual work.
  vk_pipe_run(&handle, &pipe, wg_sz);

  // Read from search output buffer.
  vk_buffer_read(&handle, &bufs[0], res, bufs[0].sz);
  vk_buffer_read(&handle, &bufs[3], search, bufs[3].sz);
}

void cleanup() {
  vk_pipe_destroy(&handle, &pipe);
  vk_buffer_destroy(&handle, &bufs[0]);
  vk_buffer_destroy(&handle, &bufs[1]);
  vk_buffer_destroy(&handle, &bufs[2]);
  vk_buffer_destroy(&handle, &bufs[3]);
  vk_handle_destroy(&handle);
}

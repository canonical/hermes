// +build ignore

#include "vmlinux.h"

#include <bpf/bpf_helpers.h>

char __license[] SEC("license") = "Dual MIT/GPL";

#define MAX_ENTRIES 1000000
#define PERF_MAX_STACK_DEPTH 127
# define PAGE_SIZE 4096

struct InfoValue {
  u64 pid;
  size_t size;
  u32 stack_id;
};

struct MemInfo {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, u64);
  __type(value, struct InfoValue);
  __uint(max_entries, MAX_ENTRIES);
};

struct MemInfo slab_info SEC(".maps");
struct MemInfo page_info SEC(".maps");

struct MemStats {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, u32);
  __type(value, size_t);
  __uint(max_entries, MAX_ENTRIES);
};

struct MemStats slab_stats SEC(".maps");
struct MemStats page_stats SEC(".maps");

struct {
  __uint(type, BPF_MAP_TYPE_STACK_TRACE);
  __uint(key_size, sizeof(u32));
  __uint(value_size, PERF_MAX_STACK_DEPTH * sizeof(u64));
  __uint(max_entries, MAX_ENTRIES);
} stack_trace SEC(".maps");

int mem_alloc(u64 addr, struct MemInfo *mem_info, struct MemStats *mem_stats, struct InfoValue *info) {
  u64 *val;

  bpf_map_update_elem(mem_info, &addr, info, BPF_NOEXIST);
  val = bpf_map_lookup_elem(mem_stats, &info->stack_id);
  if (!val) {
    u64 zero = 0;
    bpf_map_update_elem(mem_stats, &info->stack_id, &zero, BPF_NOEXIST);
    val = bpf_map_lookup_elem(mem_stats, &info->stack_id);
  }

  if (val) {
    (*val) += info->size;
  }
  return 0;
}

int mem_free(u64 addr, struct MemInfo *mem_info, struct MemStats *mem_stats) {
  struct InfoValue *info;
  u64 *val;

  info = bpf_map_lookup_elem(mem_info, &addr);
  if (!info) {
    return 0;
  }

  val = bpf_map_lookup_elem(mem_stats, &info->stack_id);
  if (!val) {
    return 0;
  }

  (*val) = ((*val) > info->size) ? (*val) - info->size : 0;
  return 0;
}

/* from /sys/kernel/debug/tracing/events/kmem/kmalloc/format */
struct SlabKmallocInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 call_site;
  void *ptr;
  size_t bytes_req;
  size_t bytes_alloc;
  u64 gfp_flags;
};

SEC("tracepoint/kmem/kmalloc")
int kmalloc(struct SlabKmallocInfo *ctx) {
  struct InfoValue info;

  __builtin_memset(&info, 0, sizeof(struct InfoValue));
  info.pid = bpf_get_current_pid_tgid();
  info.stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  info.size = ctx->bytes_alloc;

  return mem_alloc((u64)ctx->ptr, &slab_info, &slab_stats, &info);
}

/* from /sys/kernel/debug/tracing/events/kmem/kmalloc_node/format */
struct SlabKmallocNodeInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 call_site;
  void *ptr;
  size_t bytes_req;
  size_t bytes_alloc;
  u64 gfp_flags;
  int node;
};

SEC("tracepoint/kmem/kmalloc_node")
int kmalloc_node(struct SlabKmallocNodeInfo *ctx) {
  struct InfoValue info;

  __builtin_memset(&info, 0, sizeof(struct InfoValue));
  info.pid = bpf_get_current_pid_tgid();
  info.stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  info.size = ctx->bytes_alloc;
  return mem_alloc((size_t)ctx->ptr, &slab_info, &slab_stats, &info);
}

/* from /sys/kernel/debug/tracing/events/kmem/kfree/format */
struct SlabKfreeInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;

  u64 call_site;
  void *ptr;
};

SEC("tracepoint/kmem/kfree")
int kfree(struct SlabKfreeInfo *ctx) {
  return mem_free((size_t)ctx->ptr, &slab_info, &slab_stats);
}

/* from /sys/kernel/debug/tracing/events/kmem/kmem_cache_alloc/format */
struct SlabKmemCacheAllocInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 call_site;
  void *ptr;
  size_t bytes_req;
  size_t bytes_alloc;
  u64 gfp_flags;
};

SEC("tracepoint/kmem/kmem_cache_alloc")
int kmem_cache_alloc(struct SlabKmemCacheAllocInfo *ctx) {
  struct InfoValue info;

  __builtin_memset(&info, 0, sizeof(struct InfoValue));
  info.pid = bpf_get_current_pid_tgid();
  info.stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  info.size = ctx->bytes_alloc;
  return mem_alloc((size_t)ctx->ptr, &slab_info, &slab_stats, &info);
}

/* from /sys/kernel/debug/tracing/events/kmem/kmem_cache_alloc_node/format */
struct SlabKmemCacheAllocNodeInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 call_site;
  void *ptr;
  size_t bytes_req;
  size_t bytes_alloc;
  u64 gfp_flags;
  int node;
};

SEC("tracepoint/kmem/kmem_cache_alloc_node")
int kmem_cache_alloc_node(struct SlabKmemCacheAllocNodeInfo *ctx) {
  struct InfoValue info;

  __builtin_memset(&info, 0, sizeof(struct InfoValue));
  info.pid = bpf_get_current_pid_tgid();
  info.stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  info.size = ctx->bytes_alloc;
  return mem_alloc((size_t)ctx->ptr, &slab_info, &slab_stats, &info);
}

/* from /sys/kernel/debug/tracing/events/kmem/kmem_cache_free/format */
struct SlabKmemCacheFreeInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 call_site;
  void *ptr;
  char *name;
};

SEC("tracepoint/kmem/kmem_cache_free")
int kmem_cache_free(struct SlabKmemCacheFreeInfo *ctx) {
  return mem_free((size_t)ctx->ptr, &slab_info, &slab_stats);
}

/* from /sys/kernel/debug/tracing/events/kmem/mm_page_alloc/format */
struct PageAllocInfo {
  u16 common_Type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 pfn;
  u32 order;
  u64 gfp_flags;
  int migrate_type;
};

SEC("tracepoint/kmem/mm_page_alloc")
int mm_page_alloc(struct PageAllocInfo *ctx) {
  struct InfoValue info;

  __builtin_memset(&info, 0, sizeof(struct InfoValue));
  info.pid = bpf_get_current_pid_tgid();
  info.stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  info.size = (PAGE_SIZE << ctx->order);
  return mem_alloc(ctx->pfn, &page_info, &page_stats, &info);
}

/* from /sys/kernel/debug/tracing/events/kmem/mm_page_free/format */
struct PageFreeInfo {
  u16 common_type;
  u8 common_flags;
  u8 common_preempt_count;
  int common_pid;
  u64 pfn;
  int order;
};

SEC("tracepoint/kmem/mm_page_free")
int mm_page_free(struct PageFreeInfo *ctx) {
  return mem_free(ctx->pfn, &page_info, &page_stats);
}

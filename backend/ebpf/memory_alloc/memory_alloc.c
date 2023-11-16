// +build ignore

#include "vmlinux.h"

#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char __license[] SEC("license") = "Dual MIT/GPL";

#define MAX_ENTRIES 1000000
#define PERF_MAX_STACK_DEPTH 127
#define TASK_COMM_LEN 16
#define SLAB_NAME_LEN 32

#define min(a, b) ((a < b) ? a : b)

struct TaskKey {
  u64 tgid_pid;
  u64 addr;
};

struct TaskInfo {
  unsigned char slab[SLAB_NAME_LEN];
  unsigned char comm[TASK_COMM_LEN];
  size_t bytes_alloc;
  u32 stack_id;
};

struct SlabInfo {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, struct TaskKey);
  __type(value, struct TaskInfo);
  __uint(max_entries, MAX_ENTRIES);
};
struct SlabInfo slab_info SEC(".maps");

struct SlabTmpInfo {
  char slab[SLAB_NAME_LEN];
};

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, u64);
  __type(value, struct SlabTmpInfo);
  __uint(max_entries, MAX_ENTRIES);
} tgid_pid_slab SEC(".maps");

struct {
  __uint(type, BPF_MAP_TYPE_STACK_TRACE);
  __uint(key_size, sizeof(u32));
  __uint(value_size, PERF_MAX_STACK_DEPTH * sizeof(u64));
  __uint(max_entries, MAX_ENTRIES);
} stack_trace SEC(".maps");

int mem_alloc(u64 addr, u32 stack_id, const char *slab, size_t bytes_alloc) {
  u64 tgid_pid = bpf_get_current_pid_tgid();
  struct TaskKey task_key = {.tgid_pid = tgid_pid, .addr = addr};
  struct TaskInfo task_info;

  __builtin_memset(&task_info, 0, sizeof(struct TaskInfo));
  bpf_core_read_str(task_info.slab, sizeof(task_info.slab), slab);
  bpf_get_current_comm(&task_info.comm, sizeof(task_info.comm));
  task_info.bytes_alloc = bytes_alloc;
  task_info.stack_id = stack_id;

  bpf_map_update_elem(&slab_info, &task_key, &task_info, BPF_NOEXIST);

  return 0;
}

int mem_free(u64 tgid_pid, u64 addr) {
  struct TaskKey task_key = {.tgid_pid = tgid_pid, .addr = addr};

  bpf_map_delete_elem(&slab_info, &task_key);

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
  u32 stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  const char *anonymous = "anonymous";

  return mem_alloc((u64)ctx->ptr, stack_id, anonymous, ctx->bytes_alloc);
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
  u32 stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  const char *anonymous = "anonymous";

  return mem_alloc((u64)ctx->ptr, stack_id, anonymous, ctx->bytes_alloc);
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
  u64 tgid_pid = bpf_get_current_pid_tgid();

  return mem_free(tgid_pid, (u64)ctx->ptr);
}

static void set_tgid_pid_slab(struct kmem_cache *cache) {
  u64 tgid_pid = bpf_get_current_pid_tgid();
  struct SlabTmpInfo tmp_info;
  char *name = NULL;

  __builtin_memset(&tmp_info, 0, sizeof(tmp_info));
  bpf_core_read(&name, sizeof(name), &cache->name);
  bpf_core_read_str(tmp_info.slab, sizeof(tmp_info.slab), name);

  bpf_map_update_elem(&tgid_pid_slab, &tgid_pid, &tmp_info, BPF_ANY);
}

SEC("kprobe/kmem_cache_alloc")
int kmem_cache_alloc_kprobe(struct pt_regs *ctx) {
  set_tgid_pid_slab((struct kmem_cache *)PT_REGS_PARM1_CORE(ctx));

  return 0;
}

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
  u64 tgid_pid = bpf_get_current_pid_tgid();
  u32 stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  struct SlabTmpInfo *tmp_info;

  tmp_info = bpf_map_lookup_elem(&tgid_pid_slab, &tgid_pid);
  if (!tmp_info) {
    return 0;
  }

  return mem_alloc((u64)ctx->ptr, stack_id, tmp_info->slab, ctx->bytes_alloc);
}

SEC("kprobe/kmem_cache_alloc_node")
int kmem_cache_alloc_node_kprobe(struct pt_regs *ctx) {
  set_tgid_pid_slab((struct kmem_cache *)PT_REGS_PARM1_CORE(ctx));

  return 0;
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
  u32 stack_id = bpf_get_stackid(ctx, &stack_trace, 0);
  u64 tgid_pid = bpf_get_current_pid_tgid();
  struct SlabTmpInfo *tmp_info;

  tmp_info = bpf_map_lookup_elem(&tgid_pid_slab, &tgid_pid);
  if (!tmp_info) {
    return 0;
  }

  return mem_alloc((u64)ctx->ptr, stack_id, tmp_info->slab, ctx->bytes_alloc);
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
  u64 tgid_pid = bpf_get_current_pid_tgid();

  return mem_free(tgid_pid, (u64)ctx->ptr);
}

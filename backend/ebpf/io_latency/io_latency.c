// +build ignore

#include "vmlinux.h"

#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char __license[] SEC("license") = "Dual MIT/GPL";

// start track block device requests
#define DISK_NAME_LEN 16
#define TASK_COMM_LEN 32

struct blk_req_event {
    u8 disk_name[DISK_NAME_LEN];
    u8 comm[TASK_COMM_LEN];
    u32 cmd_flags;
    u64 delta_us;
    u32 pid;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, struct request *);
    __type(value, u64);
} blk_req_start_times SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} blk_req_events SEC(".maps");

// bpf2go seems to need this for generating the object
const struct blk_req_event *unused __attribute__((unused));

// start block I/O
SEC("kprobe/blk_account_io_start")
int BPF_KPROBE(kprobe__blk_account_io_start, struct request *req)
{
    u64 ts = bpf_ktime_get_ns();
    bpf_map_update_elem(&blk_req_start_times, &req, &ts, 0);
    return 0;
}


// done block I/O
SEC("kprobe/blk_account_io_done")
int BPF_KPROBE(kprobe__blk_account_io_done, struct request *req)
{
    u64 *tsp, delta_us;
    struct blk_req_event *data;
    tsp = bpf_map_lookup_elem(&blk_req_start_times, &req);
    if (!tsp) {
        return 0;   // missed issue
    }
    delta_us = (bpf_ktime_get_ns() - *tsp)/1000;
    bpf_map_delete_elem(&blk_req_start_times, &req);

    if (delta_us < 50) {
        return 0; // ignore under 50 micro seconds
    }

    data = bpf_ringbuf_reserve(&blk_req_events, sizeof(struct blk_req_event), 0);
    if (!data) {
        return 0; // couldn't reserve
    }

    data->pid = bpf_get_current_pid_tgid();
    data->delta_us = delta_us;
    bpf_core_read(&data->cmd_flags, sizeof(data->cmd_flags), &req->cmd_flags);
    // NOTE req->rq_disk may be removed in later kernels, this was created on jammy kernel
    // https://lore.kernel.org/all/20211126121802.2090656-1-hch@lst.de/
    // https://github.com/iovisor/bcc/issues/3954
    BPF_CORE_READ_STR_INTO(&data->disk_name, req, rq_disk, disk_name);
    bpf_get_current_comm(&data->comm, sizeof(data->comm));
    bpf_ringbuf_submit(data, 0);
    return 0;
}
// end track block device requests


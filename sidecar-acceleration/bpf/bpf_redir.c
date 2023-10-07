// SPDX-License-Identifier: GPL-2.0
/* Copyright (C) 2021 Intel Corporation */
/* Copyright (C) 2023 Alibaba Cloud Corporation */

/**
 * 以下代码段是从https://github.com/intel/istio-tcpip-bypass 项目下引用，版权归原作者所有
 *
 * 本项目使用了GPL 协议下的代码，遵循GPL 协议的规定发布和分发.
 *
 * 修改内容：
 * - 增加了bpf_check_is_sidecar_connection 函数来判定是否是sidecar 相关的请求
 * - 增加了bpf_map_delete_elem(&map_redir,&proxy_key) 从SOCKMAP 下快速移除非sidecar 相关socket
 * - 增加了相关debug 日志
 */


#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "bpf_sockops.h"


static inline int bpf_check_is_sidecar_connection(struct socket_4_tuple* proxy_key,struct socket_4_tuple* proxy_value) {
    if(proxy_key->local.port == 15001 || proxy_key->local.port == 15006){
        return 1;
    }
    if(proxy_value->local.port == 15001 || proxy_value->local.port == 15006){
        return 1;
    }
    return 0;
}


SEC("sk_msg")
int bpf_redir_proxy(struct sk_msg_md *msg)
{
    uint32_t rc;
    uint32_t* debug_val_ptr;
    uint32_t debug_val;
    uint32_t debug_on_index = 0;
    uint32_t debug_pckts_index = 1;
    struct socket_4_tuple proxy_key = {};
    /* for inbound traffic */
    struct socket_4_tuple key = {};
    /* for outbound and envoy<->envoy traffic*/
    struct socket_4_tuple *second_key_redir = NULL;
    sk_msg_extract4_keys(msg, &proxy_key, &key);

    uint32_t src_ip = msg->local_ip4;
    uint32_t src_port = msg->local_port;
    uint32_t dst_ip = msg->remote_ip4;
    uint32_t dst_port = bpf_ntohl(msg->remote_port);

    uint32_t log_level = 0;


    // outbound (uid 1337 or 15006 port)
    // envoy -> [ envoy -> server app]

    // outbound  (uid 1337 or 15001 port )
    // client app -> envoy

    // inbound (source or dst ip == 127.0.0.6)
    // envoy -> server app


    if (key.local.ip4 == INBOUND_ENVOY_IP || key.remote.ip4 == INBOUND_ENVOY_IP) {
        rc = bpf_msg_redirect_hash(msg, &map_redir, &key, BPF_F_INGRESS);
    } else {
        // try to delete key from  map_active_estab anyways
        bpf_map_delete_elem(&map_active_estab, &proxy_key.local);
        // outbound request
        second_key_redir = bpf_map_lookup_elem(&map_proxy, &proxy_key);

        if (second_key_redir == NULL) {
            // Non-sidecar related connections.
            bpf_map_delete_elem(&map_redir,&proxy_key);
            return SK_PASS;
        }

        if(0 == bpf_check_is_sidecar_connection(&proxy_key,second_key_redir)) {
            bpf_map_delete_elem(&map_proxy, &proxy_key);
            bpf_map_delete_elem(&map_proxy,second_key_redir);

            bpf_map_delete_elem(&map_redir,&proxy_key);
            bpf_map_delete_elem(&map_redir,second_key_redir);
            return SK_PASS;
        }

        rc = bpf_msg_redirect_hash(msg, &map_redir, second_key_redir, BPF_F_INGRESS);
    }

    debug_val_ptr = bpf_map_lookup_elem(&debug_map, &debug_on_index);
    if (debug_val_ptr) {
        log_level = *debug_val_ptr;
    }

    if(log_level > 0) {
        char src_fmt[] = "Source IP: %d.%d\n";
        bpf_trace_printk(src_fmt,sizeof(src_fmt), src_ip & 0xFF, (src_ip >> 8) & 0xFF);
        char src_fmt_1[] = ".%d.%d :%d\n";
        bpf_trace_printk(src_fmt_1,sizeof(src_fmt_1),  (src_ip >> 16) & 0xFF, (src_ip >> 24) & 0xFF,src_port);


        char dst_fmt[] = "Destination IP: %d.%d\n";
        bpf_trace_printk(dst_fmt,sizeof(dst_fmt), dst_ip & 0xFF, (dst_ip >> 8) & 0xFF);
        char dst_fmt_1[] = ".%d.%d :%d\n";
        bpf_trace_printk(dst_fmt_1,sizeof(dst_fmt_1),(dst_ip >> 16) & 0xFF, (dst_ip >> 24) & 0xFF,dst_port);
    }

    if (rc == SK_PASS) {
        if(log_level > 0){
            char fmt[] = "data redirection ok! [%x] -> [%x]\n";
            bpf_trace_printk(fmt,sizeof(fmt),proxy_key.local.ip4, proxy_key.remote.ip4);
        }

        if (log_level > 1) {
            char info_fmt[] = "data redirection succeed: [%x]->[%x],recording the number of redirections\n";
            bpf_trace_printk(info_fmt, sizeof(info_fmt), proxy_key.local.ip4, proxy_key.remote.ip4);

            debug_val_ptr = bpf_map_lookup_elem(&debug_map, &debug_pckts_index);
            if (debug_val_ptr == NULL) {
                debug_val = 0;
                debug_val_ptr = &debug_val;
            }
            __sync_fetch_and_add(debug_val_ptr, 1);
            bpf_map_update_elem(&debug_map, &debug_pckts_index, debug_val_ptr, BPF_ANY);

        }
    }else {
        if(log_level > 0){
            char info_fmt[] = "data redirection failed: [%x]->[%x]\n";
            bpf_trace_printk(info_fmt,sizeof(info_fmt),proxy_key.local.ip4, proxy_key.remote.ip4);
            char port_fmt[] = "data redirection failed port: [%d] -> [%d]\n";
            bpf_trace_printk(port_fmt,sizeof(port_fmt),proxy_key.local.port, proxy_key.remote.port);
        }
    }
    return SK_PASS;
}

char _license[] SEC("license") = "GPL";

#!/usr/bin/env bash

export redis_addr="127.0.0.1:6379"
export elastic_index="curve_cluster_alive"
export elastic_version="v0.0.1"
export start_worker="true"
export start_receiver="true"
export start_day="2022-01-01"
export debug="true"
./report

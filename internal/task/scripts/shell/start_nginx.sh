#!/usr/bin/env bash

# Created Date: 2022-08-09
# Author: Jingli Chen (Wine93)


g_listen="$1"

cat << __EOF__ > /etc/nginx/nginx.conf
daemon off;
events {
    worker_connections 768;
}
http {
    server {
        $g_listen
    }
}
__EOF__

nginx

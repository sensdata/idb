#!/bin/bash

# 获取当前用户的 Detached 会话
detached_sessions=$(screen -ls | grep Detached | awk '{print $1}')

# 循环终止每个会话
for session in $detached_sessions; do
    screen -S "$session" -X quit
done

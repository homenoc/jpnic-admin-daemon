#!/bin/zsh

max=300
for ((i = 1; i <= $max; i++)); do
  echo count:$i
  go run .
  sleep 10
done

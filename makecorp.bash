#!/bin/bash
zncprocess() {
  sed -r 's,^\[[0-9:]+] ,,g; /^\*\*\*/d; s,^<[^>]+> ,,g'
}

zncbasedir=$HOME/.znc/moddata/log
corp=/tmp/corps

for net in $zncbasedir/*; do
  for chan in $net/*; do
    for log in $chan/*; do
    done
  done
done
find $zncbasedir -type f |xargs cat |zncprocess > $corp

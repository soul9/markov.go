#!/bin/bash
cd $HOME/src/inferno-os/usr/johnny/irclog/irc.freenode.net
mkdir /tmp/nickcorpus/
for nick in uriel Kivutar Nahuel kivutar francharb alex_a nahuel Christophe_c stephane_n hek eiro matts soul9; do
  for i in *; do
    egrep '^(\+|\-) [0-9][0-9]:[0-9][0-9] *'$nick':' "$i" |sed -E 's,^(\+|\-) [0-9][0-9]:[0-9][0-9] *'$nick': ,,g'  >>/tmp/nickcorpus/$nick
  done
done

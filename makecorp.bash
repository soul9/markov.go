#!/bin/bash
mkinfernocorp() {
  usage='mkinfernocorp $nick $channel $basedir $outfile'
  if [[  -z $1 ||  -z $2 ||  -z $3 ||  -z $4 ]]; then
      echo $usage
      return
  fi
  cd $3 || ( echo "No such directory: $3" && return )
  touch $4 || ( echo "Can't create $4" && return )
  egrep '^(\+|\-) [0-9][0-9]:[0-9][0-9] *'$1':' "${2}.log" |sed -E 's,^(\+|\-) [0-9][0-9]:[0-9][0-9] *'$1': ,,g'  >> $4
}

mkznccorp() {
  usage='usage: mkznccorp $nick $channel $basedir $outfile'
  if [[  -z $1 ||  -z $2 ||  -z $3 ||  -z $4 ]]; then
    echo $usage
    return
  fi
  cd $3 || ( echo "No such directory: $3" && return )
  touch $4 || ( echo "Can't create $4" && return )
  ( for file in "${2}"*; do
    egrep -v "^\[[0-9:]+\] \*\*\*.*" $file| egrep "^\[[0-9:]+\] <$1>"  |sed -E 's,^\[[0-9:]+\] <[a-zA-Z0-9_-]+> ,,g'
  done ) >> $4
}

infernobasedir=$HOME/src/inferno-os/usr/johnny/irclog/irc.freenode.net
zncbasedir=$HOME/.znc/users/KBme/moddata/log

corpdir=/tmp/corps
mkdir $corpdir

for nick in eiro mc khatar; do
  for chan in "#biblibre" "#soul9"; do
    mkinfernocorp $nick $chan $infernobasedir $corpdir/eirocorp
    mkznccorp $nick $chan $zncbasedir $corpdir/eirocorp
  done
done

for nick in soul9 KBme; do
  for chan in "#biblibre" "#soul9"; do
    mkinfernocorp $nick $chan $infernobasedir $corpdir/soul9corp
    mkznccorp $nick $chan $zncbasedir $corpdir/soul9corp
  done
done

for nick in kivutar Kivutar kivutarrr Kivutarrr; do
  for chan in "#biblibre" "#soul9"
    mkinfernocorp $nick $chan $infernobasedir $corpdir/kivutarcorp
    mkznccorp $nick $chan $zncbasedir $corpdir/kivutarcorp
  done
done

for nick in clrh alex_a francharb nahuel; do
  mkinfernocorp $nick '#soul9' $infernobasedir $corpdir/${nick}corp
  mkznccorp $nick '#soul9' $zncbasedir $corpdir/${nick}corp
done

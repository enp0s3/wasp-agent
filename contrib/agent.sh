#!/usr/bin/bash

# Expected to be set by DS
FSROOT=${FSROOT:-/host}
DEBUG=${DEBUG}
DRY=${DRY}

_set() { echo "Setting $1 > $2" ; echo "$1" > "$2" ; }

tune_system_slice() {
  [[ -n "$SWAPPINESS" ]] && { _set "$SWAPPINESS" "$FSROOT/proc/sys/vm/swappiness"; }

  echo "Tuning system.slice"

  # Disable swap for system.slice
  _set 0 $FSROOT/sys/fs/cgroup/system.slice/memory.swap.max

  # Set latency target to protect the root slice from io trash
  MAJMIN=$(findmnt $FSROOT/ --output MAJ:MIN -n | sed "s/:.*/:0/")  # fixme can be manually provided
  echo "Using MAJMIN $MAJMIN"
  _set "$MAJMIN target=50" $FSROOT/sys/fs/cgroup/system.slice/io.latency

  echo "Tune kubepods.slice"
  MEM_HIGH_PERCENT=5
  MEM_HIGH=$(( $(< /sys/fs/cgroup/kubepods.slice/memory.max) - $(< /sys/fs/cgroup/kubepods.slice/memory.max) / $MEM_HIGH_PERCENT ))
  _set $MEM_HIGH /sys/fs/cgroup/kubepods.slice/memory.high
}

install_oci_hook() {
  # FIXME we shoud set noswap for all cgroups, not just leaves, just to be sure
  echo "installing hook"

  cp -v hook.sh $FSROOT/opt/oci-hook-swap.sh
  cp -v hook.json $FSROOT/run/containers/oci/hooks.d/swap-for-burstable.json
}

main() {
  # FIXME hardlinks are broken if FSROOT is used, but we need it
  [[ ! -d /run/containers ]] && ln -s $FSROOT/run/containers /run/containers

  tune_system_slice
  install_oci_hook

  echo "Done"

  sleep inf
}

swaptop() {
  while sleep 0.5 ; do D=$(uptime ; free -m ; find /sys/fs/cgroup -name memory.swap.current | while read FN ; do [[ -f "$FN" && "$(cat $FN)" -gt 0 ]] && { echo -n "$FN " ; numfmt --to=iec-i --suffix=B < $FN ; } ; done | sort -r -k 2 -h) ; clear ; echo "$D" | head -n 30 ; done
}

${@:-main}

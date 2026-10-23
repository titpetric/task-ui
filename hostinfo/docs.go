// Package hostinfo reads what a host of the inventory holds beyond what
// Beszel reports: its filesystems and its containers' sizes, over the
// ssh login task-ui has to it, and puts the two sources together for the
// Inventory page: the containers of a host, its durable filesystems and
// the pools of the cluster.
//
// ParseDF reads a header, then a line a filesystem of device, type, size,
// used, available, capacity and the mount point, which may hold spaces.
// Durable drops tmpfs, overlay and the other in-memory and kernel
// filesystems and docker's own mounts; of the mounts of one device, a
// bind mount among them, it keeps the one of the shortest mount point.
// BeszelFilesystems names each extra filesystem as the agent sets it.
//
// Merge orders the containers of a host by CPU, then memory, then name,
// and sets each one's share of the host's RAM. NewCluster counts the
// filesystems Beszel watches for a host without durable filesystems.
package hostinfo

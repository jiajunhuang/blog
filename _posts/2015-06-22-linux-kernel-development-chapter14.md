---
layout: post
title: "Block I/O"
tags: [linux]
---

* The difference (between character device and block device) comes down to 
whether the device accesses data randomly --- in otherwords, whether the 
device can seek to one position from another.

* The block is an abstraction of the filesystem --- filesystems can be
accessed only in multiples of a block.Although the physical device is 
addressable at the sector level, the kernel performs all disk operations in 
terms of blocks.

* the kernel (as with hardware and the sector) needs the block to be a power 
of two.The kernel also requires that a block be no larger than the page size.

* The purpose of a buffer head is to describe this mapping between the 
on-disk block and the physical in-memory buffer (which is a sequence of bytes 
on a specific page).

* the kernel does not issue block I/O requests to the disk in the order they
are received or as soon as they are received.

* Both the process scheduler and the I/O scheduler virtualize a resource 
among multiple objects.

* I/O schedulers perform two primary actions to minimize seeks: merging 
and sorting. a) Merging is the coalescing of two or more requests into one.
Consequently, merging requests reduces overhead and minimizes seeks.
b) The entire request queue is kept sorted, sectorwise, so that all seeking 
activity along the queue moves (as much as possible) sequentially over the 
sectors of the hard disk.This is similar to the algorithm employed in 
elevators ------ try to move gracefully in a single direction.

> 1. Linus Elevator: The Linus Elevator I/O scheduler performs both front and
back merging.
> 2. The Deadline I/O scheduler: ensure that write requests do not starve 
read requests.
> 3. The Anticipatory I/O scheduler aims to continue to provide excellent 
read latency, but also provide excellent global throughput.
> 4. The Complete Fair Queuing (CFQ) I/O scheduler is an I/O scheduler 
designed for specialized workloads, but that in practice actually provides 
good performance across multiple workloads.It is now the default I/O scheduler 
in Linux(2.6).
> 5. the Noop I/O Scheduler truly is a noop, merely maintaining the
request queue in near-FIFO order, from which the block device driver can pluck
requests.

# 欣赏一下K&R两位大神的代码

这段代码来自UNIX v6，作用是分配和归还内核管理的内存，使用的是first-fit算法，也就是遍历，找到第一个合适的空间，
就返回那块空间。

欣赏一下什么叫做简洁（当然，同时也就意味着阅读难度会增加）：

```c
#
/*
 */

/*
 * Structure of the coremap and swapmap
 * arrays. Consists of non-zero count
 * and base address of that many
 * contiguous units.
 * (The coremap unit is 64 bytes,
 * the swapmap unit is 512 bytes)
 * The addresses are increasing and
 * the list is terminated with the
 * first zero count.
 *
 * 一个map是一个内存单元。分两种，一种是coremap，一种是swapmap。
 * 前者64byte一个，后者512byte一个。
 */
struct map
{
	char *m_size;
	char *m_addr;
};

/*
 * Allocate size units from the given
 * map. Return the base of the allocated
 * space.
 * Algorithm is first fit.
 *
 * 使用的first fit算法，也就是遍历一遍，找到第一个合适的就返回。
 */
malloc(mp, size)
struct map *mp;  // mp是一个内存区域
{
	register int a;
	register struct map *bp;

    // 哨兵的m_size为0
	for (bp = mp; bp->m_size; bp++) {
		if (bp->m_size >= size) {  // 如果当前单元比我们要申请的更大
			a = bp->m_addr;  // 取出地址
			bp->m_addr =+ size;  // 增加地址
			if ((bp->m_size =- size) == 0)  // 把当前unit的m_size减去要申请的
				do {
					bp++;  // 往右移动一个单位
					(bp-1)->m_addr = bp->m_addr;  // 把原来的地址设置成这个单位的地址
				} while ((bp-1)->m_size = bp->m_size);  // 这里的while() 中的值，其实是 bp-m_size。也就是一直往右走，一直到哨兵
			return(a);  // 返回地址
		}
	}
	return(0);
}

/*
 * Free the previously allocated space aa
 * of size units into the specified map.
 * Sort aa into map and combine on
 * one or both ends if possible.
 */
mfree(mp, size, aa)  // aa是要归还的地址
struct map *mp;
{
	register struct map *bp;
	register int t;
	register int a;

	a = aa;
    // 从左往右找，
	for (bp = mp; bp->m_addr<=a && bp->m_size!=0; bp++);
	if (bp>mp && (bp-1)->m_addr+(bp-1)->m_size == a) {
        // if bp>mp 说明不是第一个单元
        // (bp-1)->m_addr+(bp-1)->m_size == a 如果为真，说明前一个单元的地址+大小刚好是要归还的地址
		(bp-1)->m_size =+ size;  // 把size合并到 (bp-1) 这个单元里
		if (a+size == bp->m_addr) {  // 检查 要归还的地址的右边界是不是刚好是这一个单元的起始位置
			(bp-1)->m_size =+ bp->m_size;  // 如果是的话，把这个单元和上一个单元合并
			while (bp->m_size) { // 往右依次整合map
				bp++;
				(bp-1)->m_addr = bp->m_addr;
				(bp-1)->m_size = bp->m_size;
			}
		}
	} else { // 要么是第一个单元，要么不是第一个单元但是地址不匹配
		if (a+size == bp->m_addr && bp->m_size) {  // 如果右边界是当前unit的起始地址
			bp->m_addr =- size;  // 把bp->m_addr设置成a的起始地址
			bp->m_size =+ size; // 把size加上去
		} else if (size) do { // 增加一个新的map来存储归还的内存空间
			t = bp->m_addr;
			bp->m_addr = a;
			a = t;
			t = bp->m_size;
			bp->m_size = size;
			bp++;
		} while (size = t);
	}
}
```

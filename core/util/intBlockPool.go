package util

// util/IntBlockPool.java

const INT_BLOCK_SHIFT = 13
const INT_BLOCK_SIZE = 1 << INT_BLOCK_SHIFT

/* A pool for int blocks similar to ByteBlockPool */
type IntBlockPool struct {
	buffers    [][]int
	bufferUpto int
	IntUpto    int
	Buffer     []int
	IntOffset  int
	allocator  IntAllocator
}

func NewIntBlockPool(allocator IntAllocator) *IntBlockPool {
	return &IntBlockPool{
		bufferUpto: -1,
		IntUpto:    INT_BLOCK_SIZE,
		IntOffset:  -INT_BLOCK_SIZE,
		allocator:  allocator,
	}
}

/* Expert: Resets the pool to its initial state reusing the first buffer. */
func (pool *IntBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	if pool.bufferUpto != -1 {
		// We allocated at least one buffer
		if zeroFillBuffers {
			panic("not implemented yet")
		}

		if pool.bufferUpto > 0 || !reuseFirst {
			offset := 0
			if reuseFirst {
				offset = 1
			}
			// Recycle all but the first buffer
			pool.allocator.Recycle(pool.buffers[offset : 1+pool.bufferUpto])
			for i := offset; i <= pool.bufferUpto; i++ {
				pool.buffers[i] = nil
			}
		}
		if reuseFirst {
			panic("not implemented yet")
		} else {
			pool.bufferUpto = -1
			pool.IntUpto = INT_BLOCK_SIZE
			pool.IntOffset = -INT_BLOCK_SIZE
			pool.Buffer = nil
		}
	}
}

type IntAllocator interface {
	Recycle(blocks [][]int)
}

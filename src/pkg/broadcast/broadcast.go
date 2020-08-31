package broadcast

import (
	"sync"
	"context"
)

// Broadcast to each member of a group.
//
// f(ctx, i) is called for each member i. The function f should take care to return
// immediately if ctx is cancelled.
// 
// If waitfor is 0, don't wait for members to ACK. Will always return nil error in this case.
//
// If waitfor is -1, wait for all members. Will immediately ruturn the first error, or
// nil if all members return nil.
//
func Broadcast(ctx context.Context, f func(context.Context, int) error, members, waitfor int) error {
	cancellable, cancel := context.WithCancel(ctx)
	defer cancel()
	if waitfor == 0 {
		// optimize for no-wait case
		for i := 0; i < members; i++ {
			go f(ctx, i)
		}	
		return nil
	}
	if waitfor == -1 {
		waitfor = members
	}
	var result error
	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(members)
	for i := 0; i < members; i++ {
		go func (i int) {
			if err := f(cancellable, i); err != nil {
				once.Do(func() {
					result = err
					cancel()
				})
			} else {	
				wg.Done()
			}
		} (i)
	}
	wg.Wait()
	return result
}


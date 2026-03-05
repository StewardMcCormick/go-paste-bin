package views

import (
	"context"
	"fmt"
	"sync"
	"time"

	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ViewEvent struct {
	PasteId int64
}

type viewWorker struct {
	pool           *pgxpool.Pool
	stats          map[int64]int
	mu             *sync.Mutex
	wg             *sync.WaitGroup
	buffer         chan ViewEvent
	bufferCapacity int
	isBufferFilled chan struct{}
	updateTick     time.Duration
	quite          chan struct{}
}

func NewViewsWorker(pool *pgxpool.Pool, bufferCapacity int, updateTick time.Duration) *viewWorker {
	return &viewWorker{
		pool:           pool,
		stats:          make(map[int64]int),
		mu:             &sync.Mutex{},
		wg:             &sync.WaitGroup{},
		buffer:         make(chan ViewEvent, bufferCapacity),
		bufferCapacity: bufferCapacity,
		updateTick:     updateTick,
		quite:          make(chan struct{}),
	}
}

func (w *viewWorker) Start(ctx context.Context) {
	w.startBufferMonitor(ctx)
	w.processBuffer(ctx)

	log := appctx.GetLogger(ctx)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(w.updateTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.triggerUpdate(ctx)
			case <-w.isBufferFilled:
				w.triggerUpdate(ctx)
			case <-ctx.Done():
				log.Error(ctx.Err().Error())
				return
			case <-w.quite:
				w.triggerUpdate(ctx)
				return
			}
		}
	}()
}

func (w *viewWorker) triggerUpdate(ctx context.Context) {
	log := appctx.GetLogger(ctx)
	log.Debug("trigger")
	err := w.sendUpdates(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%v - error sending update", err))
	}
}

func (w *viewWorker) SendEvent(ctx context.Context, event ViewEvent) {
	log := appctx.GetLogger(ctx)
	select {
	case w.buffer <- event:
		log.Debug(fmt.Sprintf("new event - %v", event))
	default:
		log.Info("Buffer fill, skip event")
	}
}

func (w *viewWorker) processBuffer(ctx context.Context) {
	log := appctx.GetLogger(ctx)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		for {
			select {
			case v := <-w.buffer:
				w.mu.Lock()
				w.stats[v.PasteId] += 1
				w.mu.Unlock()
			case <-w.quite:
				return
			case <-ctx.Done():
				log.Error(ctx.Err().Error())
				return
			}
		}
	}()
}

func (w *viewWorker) sendUpdates(ctx context.Context) error {
	log := appctx.GetLogger(ctx)

	w.mu.Lock()
	stats := make(map[int64]int, len(w.stats))
	for k, v := range w.stats {
		stats[k] = v
	}
	w.stats = make(map[int64]int)
	w.mu.Unlock()

	for key, value := range stats {
		query := `UPDATE paste_info SET views = views + $1 WHERE id = $2`
		_, err := w.pool.Exec(ctx, query, value, key)
		if err != nil {
			return err
		}
		log.Debug("new UPDATE request")
	}

	return nil
}

func (w *viewWorker) startBufferMonitor(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if float64(len(w.buffer))/float64(w.bufferCapacity) >= 0.8 {
					w.isBufferFilled <- struct{}{}
				}
			case <-ctx.Done():
				return
			case <-w.quite:
				return
			}
		}
	}()
}

func (w *viewWorker) Close(ctx context.Context) {
	log := appctx.GetLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	close(w.quite)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		log.Error(ctx.Err().Error())
		return
	}
}

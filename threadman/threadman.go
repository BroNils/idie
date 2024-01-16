package threadman

import (
	"idie/typed"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TaskDoneNotifier       = make(chan *Task)
	ThreadInactiveNotifier = make(chan *Threadman)
)

const (
	workerLimit = 10
)

type Option func(*Threadman)

type Threadman struct {
	//public
	ID          int
	Results     typed.Slice // []interface{}
	WorkerLimit int

	//private
	running  bool
	stopping bool

	standbyTasks typed.Slice // []func() interface{}

	standByCounter atomic.Uint64
	runningCounter atomic.Uint64
	doneCounter    atomic.Uint64

	taskCh         chan *Task
	closing        chan struct{}
	workerLimitter chan struct{}
	mutex          sync.Mutex
	wg             sync.WaitGroup

	seqTaskID int
}

func NewThreadman(fields ...Option) *Threadman {
	t := &Threadman{
		ID: 0,
	}

	t.running = false
	t.stopping = false

	t.standbyTasks = typed.Slice{}

	t.standByCounter = atomic.Uint64{}
	t.runningCounter = atomic.Uint64{}
	t.doneCounter = atomic.Uint64{}

	t.seqTaskID = 1

	t.WorkerLimit = workerLimit

	for _, field := range fields {
		field(t)
	}

	return t
}

func WithID(id int) Option {
	return func(t *Threadman) {
		t.ID = id
	}
}

func WithWorkerLimit(limit int) Option {
	return func(t *Threadman) {
		t.WorkerLimit = limit
	}
}

func (t *Threadman) worker(tParam *Task) {
	t.wg.Add(1)

	defer func() {
		<-t.workerLimitter // release some space in workerLimitter
		t.wg.Done()
	}()

	//t.DebugLogger.Printf(nil, "Thread %d worker is running task %d", t.ID, tParam.ID)

	tParam.Result = tParam.Func()
	t.runningCounter.Add(^uint64(0))
	t.doneCounter.Add(1)

	go func(tParam2 *Task) {
		select {
		case TaskDoneNotifier <- tParam2:
		}
	}(tParam)

	//t.DebugLogger.Printf(nil, "Thread %d worker is done running task %d", t.ID, tParam.ID)
}

func (t *Threadman) prepareStandbyRun() {
	//t.DebugLogger.Printf(nil, "Preparing thread %d channel", t.ID)

	if t.WorkerLimit < 1 {
		t.WorkerLimit = workerLimit
	}

	t.taskCh = make(chan *Task, t.WorkerLimit)
	t.closing = make(chan struct{})
	t.workerLimitter = make(chan struct{}, t.WorkerLimit)

	if TaskDoneNotifier == nil {
		TaskDoneNotifier = make(chan *Task)
	}
}

// StandbyRun, change state of Threadman to running and start running tasks
// please note that if you set notify to true, you should listen to TaskDoneNotifier and TaskRunNotifier
// or you will get deadlock (full channel)
func (t *Threadman) StandbyRun() {
	//t.DebugLogger.Printf(nil, "Running thread %d", t.ID)
	if t.running {
		//t.DebugLogger.Printf(nil, "Thread %d is already running", t.ID)
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.running = true
	t.prepareStandbyRun()

	//t.DebugLogger.Printf(nil, "Starting thread %d worker", t.ID)
	go func() {
		for {
			select {
			case <-t.closing:
				return
			case task, open := <-t.taskCh:
				if !open {
					//t.DebugLogger.Printf(nil, "Thread %d worker is closed", t.ID)
					return
				}

				select {
				case <-t.closing:
					return
				case t.workerLimitter <- struct{}{}:
					t.standByCounter.Add(^uint64(0))
					t.runningCounter.Add(1)

					go t.worker(task)
				}
			}
		}
	}()

	go func() {
		t.mutex.Lock()
		defer t.mutex.Unlock()

		for _, item := range t.standbyTasks.Items {
			if task, ok := item.(*Task); ok {
				go func(tParam *Task) {
					select {
					case t.taskCh <- tParam:
					}
				}(task)
			}
		}

		t.standbyTasks.Clear()
	}()

	//t.DebugLogger.Printf(nil, "Thread %d is running", t.ID)
}

func (t *Threadman) Stop() {
	/*t.mutex.Lock()
	defer t.mutex.Unlock()*/

	if !t.running {
		return
	}
	if t.stopping {
		return
	}

	t.stopping = true
	go func() {
		t.closing <- struct{}{}
		t.wg.Wait()
		close(t.workerLimitter)
		close(t.taskCh)
		close(t.closing)
		t.running = false
		t.stopping = false

		select {
		case <-time.After(3 * time.Second):
			return
		case ThreadInactiveNotifier <- t:
		}
	}()
}

func (t *Threadman) AddTask(task func() interface{}) {
	//t.DebugLogger.Printf(nil, "Adding task to thread %d", t.ID)

	if t.seqTaskID < 1 {
		t.seqTaskID = 1
	}

	taskCreated := &Task{
		ID:   t.seqTaskID,
		Func: task,
	}
	t.seqTaskID++

	if t.running {
		go func() {
			select {
			case t.taskCh <- taskCreated:
			}
		}()
	}

	if !t.running {
		t.standbyTasks.Append(taskCreated)
		t.standByCounter.Add(1)
	}

	//t.DebugLogger.Printf(nil, "Task added to thread %d with id %d", t.ID, taskCreated.ID)
}

func (t *Threadman) IsRunning() bool {
	return t.running
}

func (t *Threadman) IsStopping() bool {
	return t.stopping
}

func (t *Threadman) GetStandByCounter() uint64 {
	return t.standByCounter.Load()
}

func (t *Threadman) GetRunningCounter() uint64 {
	return t.runningCounter.Load()
}

func (t *Threadman) GetDoneCounter() uint64 {
	return t.doneCounter.Load()
}

func (t *Threadman) AddStandByCounter(i uint64) {
	t.doneCounter.Add(i)
}

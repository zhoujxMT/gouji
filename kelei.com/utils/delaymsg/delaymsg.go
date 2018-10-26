package delaymsg

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	slotsCount = 36000
)

var (
	taskLoopInterval, _ = time.ParseDuration("0.1s")
)

const (
	STATUS_PAUSED = iota
	STATUS_RUNNING
)

//延迟消息
type DelayMessage struct {
	//当前下标
	curIndex int
	//环形槽
	slots [slotsCount]map[string]*Task
	//状态
	status int
	//关闭
	closed chan bool
	//任务关闭
	taskClose chan bool
	//时间关闭
	timeClose chan bool
	//启动时间
	startTime time.Time
	//任务映射环形槽
	taskMapSlot map[string]int64
	//锁
	lock sync.Mutex
	//累计毫秒数
	millisecond int
}

//执行的任务函数
type TaskFunc func(args ...interface{})

//任务
type Task struct {
	//循环次数
	cycleNum int
	//执行的函数
	exec   TaskFunc
	params []interface{}
}

//创建一个延迟消息
func NewDelayMessage() *DelayMessage {
	dm := &DelayMessage{
		curIndex:    0,
		status:      STATUS_RUNNING,
		closed:      make(chan bool),
		taskClose:   make(chan bool),
		timeClose:   make(chan bool),
		startTime:   time.Now(),
		taskMapSlot: make(map[string]int64),
	}
	dm.lock.Lock()
	defer dm.lock.Unlock()
	for i := 0; i < slotsCount; i++ {
		dm.slots[i] = make(map[string]*Task)
	}
	return dm
}

func (dm *DelayMessage) getStatus() int {
	return dm.status
}

func (dm *DelayMessage) setStatus(status int) {
	dm.status = status
}

//启动延迟消息
func (dm *DelayMessage) Start() {
	//	go dm.taskLoop()
	go dm.timeLoop()
	go func() {
		select {
		case <-dm.closed:
			{
				dm.timeClose <- true
				break
			}
		}
	}()
}

//暂停延迟消息
func (dm *DelayMessage) Pause() {
	dm.setStatus(STATUS_PAUSED)
}

//恢复延迟消息
func (dm *DelayMessage) Resume() {
	dm.setStatus(STATUS_RUNNING)
}

//关闭延迟消息
func (dm *DelayMessage) Close() {
	dm.closed <- true
}

//处理每0.1秒的任务
func (dm *DelayMessage) taskLoop_back() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	//取出当前的槽的任务
	tasks := dm.slots[dm.curIndex]
	if len(tasks) > 0 {
		//遍历任务，判断任务循环次数等于0，则运行任务
		//否则任务循环次数减1
		for k, v := range tasks {
			if v.cycleNum == 0 {
				go func() {
					v.exec(v.params...)
					//删除运行过的任务
					delete(tasks, k)
				}()
			} else {
				v.cycleNum--
			}
		}
	}
}

//处理每0.1秒移动下标
func (dm *DelayMessage) timeLoop() {
	defer func() {
		fmt.Println("timeLoop exit")
	}()
	tick := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-dm.timeClose:
			{
				return
			}
		case <-tick.C:
			{
				if dm.getStatus() == STATUS_RUNNING {
					//判断当前下标，如果等于3599则重置为0，否则加1
					if dm.curIndex == slotsCount-1 {
						dm.curIndex = 0
					} else {
						dm.curIndex++
					}
					dm.taskLoop_back()
				}
			}
		}
	}
}

//添加任务
func (dm *DelayMessage) AddTask(t time.Time, key string, exec TaskFunc, params []interface{}) error {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	if dm.startTime.After(t) {
		return errors.New("时间错误")
	}
	//当前时间与指定时间相差秒数
	//	subSecond := t.UnixNano()/1e8 - dm.startTime.UnixNano()/1e8
	subSecond := t.Sub(dm.startTime).Nanoseconds() / 1e6 / 100
	//计算循环次数
	cycleNum := int(subSecond / slotsCount)
	//计算任务所在的slots的下标
	ix := subSecond % slotsCount
	//把任务加入tasks中
	tasks := dm.slots[ix]
	if _, ok := tasks[key]; ok {
		return errors.New("该slots中已存在key为" + key + "的任务")
	}
	dm.taskMapSlot[key] = ix
	tasks[key] = &Task{
		cycleNum: cycleNum,
		exec:     exec,
		params:   params,
	}
	return nil
}

//暂停任务
func (dm *DelayMessage) PauseTask() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.Pause()
}

//恢复任务
func (dm *DelayMessage) ResumeTask() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.Resume()
}

//删除任务
func (dm *DelayMessage) RemoveTask(key string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	ix := dm.taskMapSlot[key]
	if ix == 0 {
		return
	}
	tasks := dm.slots[ix]
	delete(tasks, key)
}

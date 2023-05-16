# Task Queue

Task is the interface to the smallest unit of SP background service interaction. 
Task scheduling and execution are directly related to the order of task arrival, 
so task queue is a relatively important basic interface used by all modules inside 
SP.

# Concept

## Task Queue With Limit

Task execution needs to consume certain resources. Different task types have large 
differences in Memory, Bandwidth, and CPU consumption. The available resources of the 
nodes executing the task are uneven. Therefore, resources need to be considered when 
scheduling tasks. The `Task Queue With Limit` is to consider resources.

## Task Queue Strategy

Conventional queues cannot fully meet the requirements of tasks. For example, the 
retired strategy of tasks inside the queue, when the conventional queue is full, 
it cannot be pushed any more, however, tasks that fail after retries may need to be 
retired. For different types of task retired and pick up, etc. the strategies are 
different, the `Task Queue Strategy` is an interface that supports custom strategies.

# Task Queue Types

## TQueue

TQueue is the interface to task queue. The task queue is mainly used to maintain tasks 
are running. In addition to supporting conventional FIFO operations, task queue also 
has some customized operations for task. For example, Has, PopByKey.

## TQueueWithLimit

TQueueWithLimit is the interface task queue that takes resources into account. Only 
tasks with less than required resources can be popped out.

## TQueueOnStrategy

TQueueOnStrategy is a combination of TQueue and TQueueStrategy， it is the interface to 
task queue and the queue supports customize strategies to filter task for popping and 
retiring task.

## TQueueOnStrategyWithLimit

TQueueOnStrategyWithLimit is a combination of TQueueWithLimit and TQueueStrategy，it is 
the interface to task queue that takes resources into account, and the queue supports 
customize strategies to filter task for popping and retiring task.
# ResourceManager

ResourceManager is the interface to the resource management subsystem.
The ResourceManager tracks and accounts for resource usage in the stack,
from the internals to the application, and provides a mechanism to limit
resource usage according to a user configurable policy.

# Concept

## Limit

The modus operandi of the resource manager is to restrict resource usage 
at the time of reservation. When a component of the stack needs to use a 
resource, it reserves it in the appropriate scope. The resource manager 
gates the reservation against the scope applicable limits; if the limit is 
exceeded, then an error is up the component to act accordingly. At the lower 
levels of the stack, this will normally signal a failure of some sorts, like 
failing to opening a connection, which will propagate to the programmer. Some 
components may be able to handle resource reservation failure more gracefully.

The `Limit` is configured by the user policy. supports `MemoryLimit`,`ConnectionsLimit`, 
`FDLimit(file descriptor limit)`, `TaskLimit`. The `MemoryLimit` limits maximum memory 
for a component reservation. The `ConnectionsLimit` limits maximum connections number 
for a component reservation, includes inbound and outbound connections. The `FDLimit` 
limits maximum fd number for a component reservation, includes the open of sockets and 
files.

The `TaskLimit` is unique to SP. Task is the smallest unit of SP internal component 
interaction. Each component can limit the number of tasks executed. Tasks are divided 
into high, medium and low priorities, the priority can be used as an important basis 
for task scheduling within the SP. The higher the priority, the faster it is expected 
to be executed, and the resources will be assigned priority for execution, for example: 
seal object. The lower the priority, it can be executed later, and the resource 
requirements are not so urgent, for example: delayed deletion.

## Scope

Resource Management through the ResourceManager is based on the concept of Resource 
Management Scopes, whereby resource usage is constrained by a DAG of scopes, The following 
diagram illustrates the structure of the resource constraint DAG:
```
System(Topmost Scope)

	+------------> Transient(Scope)........................+................+
	|                                                      .                .
	+------------> Service(Scope)------------------------- . ----------+    .
	|                                                      .           |    .
			   	   +--->  Connection/Memory/Task---------- . ----------+    .
```

Scope is an important node in DAG, Scope has children and siblings scopes. There is a 
directed edge between them. The children scopes share the limit of the parent scope, the 
child scope reserves the resources, the parent scope will reduce the corresponding amount 
of resources. Sibling scopes also have directionality, for example Service A Scope depends 
on(points to) System Scope, Service A reserves the resources, the System Scope will reduce 
the corresponding amount of resources. On the contrary, if the System Scope reserves resources,
it will not affect Service A.

# Example
```go
    rcmgr := &ResourceManager{}
    serviceScope, err := rcmgr.OpenService(...)
	if err != nil { ... }
    defer scope.Close()
	
    s, err := serviceScope.BeginSpan()
	if err != nil { ... }
	defer s.Done()

	if err := s.ReserveMemory(...); err != nil { ... }
	// ... use memory
```
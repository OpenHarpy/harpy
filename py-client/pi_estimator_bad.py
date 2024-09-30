from sdk import (
    Session,
    MapTask,
    ReduceTask,
    TransformTask,
    Result,
    TaskSetResults
)

def map_fun(i: int, n: int) -> float:
    return 4 / (1 + ((i-0.5)/n)**2)

def reduce_fun(*results: float, n:int) -> float:
    sum_all_members = 0
    for result in results:
        sum_all_members += result
    return 1/n * sum_all_members

session = Session().create_session()
taskSet = session.create_task_set()

N = 100

def make_map_task(i, n):
    return MapTask(name="map", fun=map_fun, args=[], kwargs={"i": i, "n": n})
reduce_task = ReduceTask(name="reduce", fun=reduce_fun, args=[], kwargs={"n": N})

taskSet.add_maps([make_map_task(i, N) for i in range(1, N+1)])
taskSet.add_reduce(reduce_task)

taskSetResult:TaskSetResults = taskSet.execute()

print(taskSetResult)

session.close()
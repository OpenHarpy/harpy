from harpy import (
    Session,
    MapTask,
    ReduceTask,
    TaskSetResults
)

def map_fun(i_start:int, i_end:int, n:int) -> float:
    values = [
        4 / (1 + ((i-0.5)/n)**2)
        for i in range(i_start, i_end+1)
    ]
    return 1/n * sum(values)

def reduce_fun(*results: float) -> float:
    return sum(results)

session = Session().create_session()
taskSet = session.create_task_set()

N = 1000
taskSet.add_maps([
    MapTask(name="map", fun=map_fun, args=[], kwargs={"i_start": 1, "i_end": 250, "n": N}),
    MapTask(name="map", fun=map_fun, args=[], kwargs={"i_start": 251, "i_end": 500, "n": N}),
    MapTask(name="map", fun=map_fun, args=[], kwargs={"i_start": 501, "i_end": 750, "n": N}),
    MapTask(name="map", fun=map_fun, args=[], kwargs={"i_start": 751, "i_end": 1000, "n": N})
])
taskSet.add_reduce(
    ReduceTask(name="reduce", fun=reduce_fun, args=[], kwargs={})
)

taskSetResult:TaskSetResults = taskSet.execute()
print(taskSetResult)
session.close()
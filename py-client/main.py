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

def reduce_fun(*results: int, n:int) -> int:
    sum_all_members = 0
    for result in results:
        sum_all_members += result
    return 1/n * sum_all_members

def transform_fun(result: int) -> int:
    return result * 2

session = Session().create_session()
taskSet = session.create_task_set()

map_task_1 = MapTask(name="map", fun=map_fun, args=[], kwargs={"number": 2})
map_task_2 = MapTask(name="map", fun=map_fun, args=[], kwargs={"number": 4})
map_task_3 = MapTask(name="map", fun=map_fun, args=[], kwargs={"number": 6})

reduce_task = ReduceTask(name="reduce", fun=reduce_fun, args=[], kwargs={})
transform_task = TransformTask(name="transform", fun=transform_fun, args=[], kwargs={})

taskSet.add_maps([map_task_1, map_task_2, map_task_3])
taskSet.add_reduce(reduce_task)
taskSet.add_transform(transform_task)

taskSetResult:TaskSetResults = taskSet.execute()

print(taskSetResult)

session.close()
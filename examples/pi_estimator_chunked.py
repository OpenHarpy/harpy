import time

from harpy import (
    Session,
    BatchMapTask,
    MapTask,
    ReduceTask,
    TransformTask,
    Result,
    TaskSetResults
)

def map_fun(i: int, n: int) -> float:
    return 4 / (1 + ((i-0.5)/n)**2)

def reduce_fun(*results: list[float], n:int) -> float:
    sum_all_members = 0
    for sub_result in results:
        for result in sub_result:
            sum_all_members += result
    return 1/n * sum_all_members

start_time = time.time()
print("Creating session... (this can take a while because of isolated environment setup)")
session = Session()
print(f"Session creation time: {time.time() - start_time:.4f} seconds")

taskSet = session.create_task_set()

N = 500

def make_map_task(i, n):
    return MapTask(name="map", fun=map_fun, args=[], kwargs={"i": i, "n": n})

reduce_task = ReduceTask(name="reduce", fun=reduce_fun, args=[], kwargs={"n": N})

bmt = BatchMapTask(
    name="batch_map",
    batch_size=50,
    map_tasks=[make_map_task(i, N) for i in range(1, N+1)]
)

taskSet.add_batch_maps(bmt)
taskSet.add_reduce(reduce_task)

start_time = time.time()
taskSetResult:TaskSetResults = taskSet.collect()
print(f"Task set execution time: {time.time() - start_time:.4f} seconds")

print(taskSetResult)

session.close()
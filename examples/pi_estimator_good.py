import time

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

# Measure session creation time
start_time = time.time()
print("Creating session... (this can take a while because of isolated environment setup)")
session = Session().create_session()
print(f"Session creation time: {time.time() - start_time:.4f} seconds")

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

# Measure task set execution time
start_time = time.time()
taskSetResult:TaskSetResults = taskSet.collect()
print(f"Task set execution time: {time.time() - start_time:.4f} seconds")

print(taskSetResult)

# Measure session close time
start_time = time.time()
session.close()
print(f"Session close time: {time.time() - start_time:.4f} seconds")